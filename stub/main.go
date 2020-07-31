package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	platform "github.com/shanegibbs/toolbox/toolbox"
)

func dockerCli() *client.Client {
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Fatalf("Failed to create docker client: %v", err)
	}
	return cli
}

func getToolboxID(cli *client.Client) *string {
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		for _, name := range container.Names {
			if name == "/toolbox" {
				return &container.ID
			}
		}
	}

	return nil
}

func setupToolbox(cli *client.Client) *string {
	options := platform.BuildInitOptions()

	var config container.Config
	var hostConfig container.HostConfig
	var networkingConfig network.NetworkingConfig

	config.Image = "toolbox"
	config.Cmd = []string{"tail -f /dev/null"}
	config.Env = []string{
		"TOOLBOX_INIT_WAIT=true",
		fmt.Sprintf("TOOLBOX_INIT_OPTIONS=%s", options.AsString()),
	}
	config.Hostname = "toolbox"
	config.Tty = true

	hostConfig.NetworkMode = "host"
	// hostConfig.AutoRemove = true
	hostConfig.Privileged = true

	hostConfig.Mounts = []mount.Mount{}
	hostConfig.Mounts = append(hostConfig.Mounts, cloneFromHost(options.Home))
	hostConfig.Mounts = append(hostConfig.Mounts, cloneFromHost("/var/run/docker.sock"))
	hostConfig.Mounts = append(hostConfig.Mounts, cloneFromHost("/run/host-services/ssh-auth.sock"))
	hostConfig.Mounts = append(hostConfig.Mounts, bindIntoLocal("/tmp"))
	hostConfig.Mounts = append(hostConfig.Mounts, bindIntoLocal("/Users"))

	create, err := cli.ContainerCreate(context.Background(), &config, &hostConfig, &networkingConfig, "toolbox")
	if err != nil {
		log.Fatalf("Failed to create container: %v", err)
	}

	// log.Printf("created %v", create.ID)

	err = cli.ContainerStart(context.Background(), create.ID, types.ContainerStartOptions{})
	if err != nil {
		log.Fatalf("Failed to start container: %v", err)
	}

	// log.Printf("started %v", create.ID)

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
		// log.Println("stdin is from pipe")
		return false
	} else {
		// log.Println("stdin is from terminal")
		return true
	}
}

func runToolbox() {
	options := platform.BuildRunOptions()

	var args []string
	args = append(args, "docker", "exec")
	args = append(args, "-i")
	if inTerminal() {
		args = append(args, "-t")
	}
	args = append(args, "--env", "TOOLBOX_RUN_OPTIONS")
	args = append(args, "toolbox")
	args = append(args, "/toolbox-run")

	os.Setenv("TOOLBOX_RUN_OPTIONS", options.AsString())

	// hand proc off to docker
	// log.Print("Handoff to docker")
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

func main() {
	cli := dockerCli()

	id := getToolboxID(cli)
	if id == nil {
		id = setupToolbox(cli)
	}

	runToolbox()

	// quickRun()
}
