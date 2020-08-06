package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	platform "github.com/shanegibbs/toolbox/toolbox"
	log "github.com/sirupsen/logrus"
)

func dockerCli() *client.Client {
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Fatalf("Failed to create docker client: ", err)
	}
	return cli
}

func getToolboxID(ctx context.Context, cli *client.Client) *string {
	listOptions := types.ContainerListOptions{}
	listOptions.All = true
	containers, err := cli.ContainerList(ctx, listOptions)
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		for _, name := range container.Names {
			if name == "/toolbox" {
				log.Debug("Found container ID: ", container.ID)
				return &container.ID
			}
		}
	}

	return nil
}

func setupToolbox(ctx context.Context, cli *client.Client) *string {
	log.Trace("starting new toolbox")

	options := platform.BuildInitOptions()

	var config container.Config
	var hostConfig container.HostConfig
	var networkingConfig network.NetworkingConfig

	config.Image = "toolboxed"
	config.Cmd = []string{"tail -f /dev/null"}
	config.Env = []string{
		"TOOLBOX_INIT_WAIT=true",
		fmt.Sprintf("TOOLBOX_INIT_OPTIONS=%s", options.AsString()),
	}
	config.Hostname = "toolbox"
	config.Domainname = "local"
	config.Tty = true

	hostConfig.NetworkMode = "host"
	hostConfig.ExtraHosts = []string{
		"toolbox:127.0.1.1",
		"toolbox.local:127.0.1.1",
	}
	// hostConfig.AutoRemove = true
	hostConfig.Privileged = true

	hostConfig.Mounts = []mount.Mount{}
	hostConfig.Mounts = append(hostConfig.Mounts, cloneFromHost(options.Home))
	hostConfig.Mounts = append(hostConfig.Mounts, cloneFromHost("/var/run/docker.sock"))
	hostConfig.Mounts = append(hostConfig.Mounts, cloneFromHost("/run/host-services/ssh-auth.sock"))
	hostConfig.Mounts = append(hostConfig.Mounts, bindIntoLocal("/tmp"))
	hostConfig.Mounts = append(hostConfig.Mounts, bindIntoLocal("/Users"))

	create, err := cli.ContainerCreate(ctx, &config, &hostConfig, &networkingConfig, "toolbox")
	if err != nil {
		log.Fatal("Failed to create container: ", err)
	}

	log.Debug("created ", create.ID)

	err = cli.ContainerStart(ctx, create.ID, types.ContainerStartOptions{})
	if err != nil {
		log.Fatal("Failed to start container: ", err)
	}

	log.Debug("started ", create.ID)

	var logOptions types.ContainerLogsOptions
	logOptions.Follow = false
	logOptions.ShowStdout = true
	logOptions.ShowStderr = true
	logOptions.Tail = "100"

	logStream, err := cli.ContainerLogs(ctx, create.ID, logOptions)
	buf := new(strings.Builder)
	_, err = io.Copy(buf, logStream)
	if err != nil {
		log.Fatal("Failed to read logs: ", err)
	}

	log.Debug("Logs: ", buf.String())

	return &create.ID
}

func cloneFromHost(path string) mount.Mount {
	return mount.Mount{Type: mount.TypeBind, Source: path, Target: path}
}

func bindIntoLocal(path string) mount.Mount {
	return mount.Mount{Type: mount.TypeBind, Source: path, Target: fmt.Sprintf("/host", path)}
}

func inTerminal() bool {
	fi, _ := os.Stdin.Stat()
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		log.Debug("stdin is from pipe")
		return false
	} else {
		log.Debug("stdin is from terminal")
		return true
	}
}

func runToolbox() {
	initOptions := platform.BuildInitOptions()
	runOptions := platform.BuildRunOptions()

	var args []string
	args = append(args, "docker", "exec")
	args = append(args, "-i")
	if inTerminal() {
		args = append(args, "-t")
	}
	args = append(args, "--env", "SHAM_INIT_OPTIONS")
	args = append(args, "--env", "SHAM_RUN_OPTIONS")
	args = append(args, "toolbox")
	args = append(args, "/sham-run")

	os.Setenv("SHAM_INIT_OPTIONS", initOptions.AsString())
	os.Setenv("SHAM_RUN_OPTIONS", runOptions.AsString())

	// hand proc off to docker
	log.Debug("Handing off to docker now")
	if err := syscall.Exec("/usr/local/bin/docker", args, os.Environ()); err != nil {
		log.Fatal(err)
	}
}

func quickRun() {
	options := platform.BuildInitOptions()

	var args []string
	args = append(args, "docker", "run")
	args = append(args, "-it")
	args = append(args, "--env", "TOOLBOX_OPTIONS")
	args = append(args, "--hostname", "toolbox")
	args = append(args, "--mount", "type=bind,src=/run/host-services/ssh-auth.sock,target=/run/host-services/ssh-auth.sock")
	args = append(args, "-v", fmt.Sprintf("%s:%s", options.Home, options.Home))
	args = append(args, "-v", "/private:/private")
	args = append(args, "-v", "/var/run/docker.sock:/var/run/docker.sock")
	args = append(args, "-v", "/Users:/host/Users")
	args = append(args, "-v", "/tmp:/host/tmp")
	args = append(args, "-v", "/Volumes:/host/Volumes")
	args = append(args, "platform")

	os.Setenv("TOOLBOX_OPTIONS", options.AsString())

	// hand proc off to docker
	if err := syscall.Exec("/usr/local/bin/docker", args, os.Environ()); err != nil {
		log.Fatal(err)
	}
}

func buildToolboxImage(ctx context.Context, cli *client.Client) {
	options := platform.BuildInitOptions()
	setupOptions := options.AsString()

	log.Debug("SHAM_INIT_OPTIONS: ", setupOptions)

	imageArg := "ubuntu:latest"
	userID := fmt.Sprintf("%v", options.Uid)

	var buildOptions types.ImageBuildOptions
	buildOptions.RemoteContext = "http://localhost/Dockerfile"
	buildOptions.Remove = true
	buildOptions.ForceRemove = true
	buildOptions.NoCache = true
	buildOptions.Tags = []string{"toolboxed:latest"}
	buildOptions.BuildArgs = make(map[string]*string)
	buildOptions.BuildArgs["IMAGE"] = &imageArg
	buildOptions.BuildArgs["USER_ID"] = &userID
	buildOptions.BuildArgs["SHAM_INIT_OPTIONS"] = &setupOptions

	log.Debug("building image")

	resp, err := cli.ImageBuild(ctx, nil, buildOptions)
	if err != nil {
		log.Fatal("Failed to build toolbox: ", err)
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		log.Debug(line)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal("reading build output: ", err)
	}

	log.Debug("image build complete")
}

func main() {
	platform.SetupLogging()

	log.Trace("Start. Args: ", os.Args)

	ctx := context.Background()

	cli := dockerCli()

	id := getToolboxID(ctx, cli)

	if id == nil {
		buildToolboxImage(ctx, cli)
		id = setupToolbox(ctx, cli)
	}

	runToolbox()

	// quickRun()
}
