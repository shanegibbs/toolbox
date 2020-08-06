package sham

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
	log "github.com/sirupsen/logrus"
)

type Sham struct {
	ctx         context.Context
	l           *log.Entry
	initOptions *InitOptions
	runOptions  *RunOptions
	docker      *client.Client
}

func New() *Sham {
	log.Trace("Start. Args: ", os.Args)
	return &Sham{
		ctx: context.Background(),
		l:   log.WithField("prefix", "foo"),
	}
}

func (sham *Sham) BuildInitOptions() {
	sham.initOptions = BuildInitOptions()
}

func (sham *Sham) BuildRunOptions() {
	sham.runOptions = BuildRunOptions()
}

func (sham *Sham) CreateDockerClient() {
	docker, err := client.NewEnvClient()
	if err != nil {
		log.Fatalf("Failed to create docker client: ", err)
	}
	sham.l.Debug("Created docker client")
	sham.docker = docker
}

func (sham *Sham) FindContainerID() *string {
	listOptions := types.ContainerListOptions{}
	listOptions.All = true
	containers, err := sham.docker.ContainerList(sham.ctx, listOptions)
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

func (sham *Sham) CreateContainer() *string {
	log.Trace("starting new container")

	var config container.Config
	var hostConfig container.HostConfig
	var networkingConfig network.NetworkingConfig

	config.Image = "toolboxed"
	config.Cmd = []string{"tail -f /dev/null"}
	// config.Env = []string{
	// 	"TOOLBOX_INIT_WAIT=true",
	// 	fmt.Sprintf("TOOLBOX_INIT_OPTIONS=%s", sham.initOptions.AsString()),
	// }
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
	hostConfig.Mounts = append(hostConfig.Mounts, cloneFromHost(sham.initOptions.Home))
	hostConfig.Mounts = append(hostConfig.Mounts, cloneFromHost("/var/run/docker.sock"))
	hostConfig.Mounts = append(hostConfig.Mounts, cloneFromHost("/run/host-services/ssh-auth.sock"))
	hostConfig.Mounts = append(hostConfig.Mounts, bindIntoLocal("/tmp"))
	hostConfig.Mounts = append(hostConfig.Mounts, bindIntoLocal("/Users"))

	create, err := sham.docker.ContainerCreate(sham.ctx, &config, &hostConfig, &networkingConfig, "toolbox")
	if err != nil {
		log.Fatal("Failed to create container: ", err)
	}

	log.Debug("created ", create.ID)

	err = sham.docker.ContainerStart(sham.ctx, create.ID, types.ContainerStartOptions{})
	if err != nil {
		log.Fatal("Failed to start container: ", err)
	}

	log.Debug("started ", create.ID)

	var logOptions types.ContainerLogsOptions
	logOptions.Follow = false
	logOptions.ShowStdout = true
	logOptions.ShowStderr = true
	logOptions.Tail = "100"

	logStream, err := sham.docker.ContainerLogs(sham.ctx, create.ID, logOptions)
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

func (sham *Sham) SendCommandToContainer() {
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

	os.Setenv("SHAM_INIT_OPTIONS", sham.initOptions.AsString())
	os.Setenv("SHAM_RUN_OPTIONS", sham.runOptions.AsString())

	// hand proc off to docker
	log.Debug("Handing off to docker now")
	if err := syscall.Exec("/usr/local/bin/docker", args, os.Environ()); err != nil {
		log.Fatal(err)
	}
}

func (sham *Sham) BuildImage() {
	setupOptions := sham.initOptions.AsString()

	sham.l.Debug("SHAM_INIT_OPTIONS: ", setupOptions)

	imageArg := "ubuntu:latest"
	userID := fmt.Sprintf("%v", sham.initOptions.Uid)

	var buildOptions types.ImageBuildOptions
	buildOptions.RemoteContext = "http://localhost/Dockerfile"
	// buildOptions.Remove = true
	// buildOptions.ForceRemove = true
	// buildOptions.NoCache = true
	buildOptions.Tags = []string{"toolboxed:latest"}
	buildOptions.BuildArgs = make(map[string]*string)
	buildOptions.BuildArgs["IMAGE"] = &imageArg
	buildOptions.BuildArgs["USER_ID"] = &userID
	buildOptions.BuildArgs["SHAM_INIT_OPTIONS"] = &setupOptions

	log.Debug("building image")

	resp, err := sham.docker.ImageBuild(sham.ctx, nil, buildOptions)
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

	sham.l.Debug("image build complete")
}

func CmdSham() {
	sham := New()
	sham.CreateDockerClient()
	sham.BuildInitOptions()
	sham.BuildRunOptions()

	id := sham.FindContainerID()

	if id == nil {
		sham.BuildImage()
		id = sham.CreateContainer()
	}

	sham.SendCommandToContainer()
}
