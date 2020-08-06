package sham

import (
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
)

func (sham *Sham) CreateContainer() {
	sham.l.Debug("starting new container")

	var config container.Config
	var hostConfig container.HostConfig
	var networkingConfig network.NetworkingConfig

	if sham.shamImage == nil {
		sham.l.Fatal("sham image not found")
	}

	config.Image = sham.shamImage.ID
	// config.Cmd = []string{"tail -f /dev/null"}
	// config.Env = []string{
	// 	"TOOLBOX_INIT_WAIT=true",
	// 	fmt.Sprintf("TOOLBOX_INIT_OPTIONS=%s", sham.initOptions.AsString()),
	// }
	config.Hostname = sham.config.Name
	config.Domainname = "local"
	config.Tty = true

	hostConfig.NetworkMode = "host"
	hostConfig.ExtraHosts = []string{
		fmt.Sprintf("%s:127.0.1.1", sham.config.Name),
		fmt.Sprintf("%s.local:127.0.1.1", sham.config.Name),
	}
	// hostConfig.AutoRemove = true
	hostConfig.Privileged = true

	hostConfig.Mounts = []mount.Mount{}
	hostConfig.Mounts = append(hostConfig.Mounts, cloneFromHost(sham.initOptions.Home))
	hostConfig.Mounts = append(hostConfig.Mounts, cloneFromHost("/var/run/docker.sock"))
	hostConfig.Mounts = append(hostConfig.Mounts, cloneFromHost("/run/host-services/ssh-auth.sock"))
	hostConfig.Mounts = append(hostConfig.Mounts, bindIntoLocal("/tmp"))
	hostConfig.Mounts = append(hostConfig.Mounts, bindIntoLocal("/Users"))

	create, err := sham.docker.ContainerCreate(sham.ctx, &config, &hostConfig, &networkingConfig, "")
	if err != nil {
		sham.l.Fatal("Failed to create container: ", err)
	}

	sham.l.Debug("created ", create.ID)

	err = sham.docker.ContainerStart(sham.ctx, create.ID, types.ContainerStartOptions{})
	if err != nil {
		sham.l.Fatal("Failed to start container: ", err)
	}

	sham.l.Debug("started ", create.ID)

	var logOptions types.ContainerLogsOptions
	logOptions.Follow = false
	logOptions.ShowStdout = true
	logOptions.ShowStderr = true
	logOptions.Tail = "100"

	logStream, err := sham.docker.ContainerLogs(sham.ctx, create.ID, logOptions)
	buf := new(strings.Builder)
	_, err = io.Copy(buf, logStream)
	if err != nil {
		sham.l.Fatal("Failed to read logs: ", err)
	}

	sham.l.Debug("Logs: ", buf.String())

	listOptions := types.ContainerListOptions{}
	listOptions.All = true
	listOptions.Filters = filters.NewArgs()
	listOptions.Filters.Add("label", fmt.Sprintf("%s=%s", keyShamName, sham.config.Name))
	listOptions.Filters.Add("label", fmt.Sprintf("%s=%s", keyShamImageRef, sham.config.Image))
	containers, err := sham.docker.ContainerList(sham.ctx, listOptions)
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		for _, name := range container.Names {
			sham.l.Debug("Found sham container ID: ", container.ID, " with name: ", name, " ", container.State, "/", container.Status)
		}
		if container.ID == create.ID {
			sham.shamContainer = &container
			return
		}
	}
}

func cloneFromHost(path string) mount.Mount {
	return mount.Mount{Type: mount.TypeBind, Source: path, Target: path}
}

func bindIntoLocal(path string) mount.Mount {
	return mount.Mount{Type: mount.TypeBind, Source: path, Target: fmt.Sprintf("/host", path)}
}
