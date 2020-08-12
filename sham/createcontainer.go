package sham

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/moby/term"
	"github.com/shanegibbs/toolbox/sham/ext/streams"
)

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (sham *Sham) BuildContainerConfig() (*container.Config, *container.HostConfig, *network.NetworkingConfig) {
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
	hostConfig.AutoRemove = true
	// hostConfig.Privileged = true

	hostConfig.Mounts = []mount.Mount{}
	hostConfig.Mounts = append(hostConfig.Mounts, cloneFromHost(sham.initOptions.Home))
	if pathExists("/var/run/docker.sock") {
		hostConfig.Mounts = append(hostConfig.Mounts, cloneFromHost("/var/run/docker.sock"))
	}
	if pathExists("/run/host-services/ssh-auth.sock") {
		hostConfig.Mounts = append(hostConfig.Mounts, cloneFromHost("/run/host-services/ssh-auth.sock"))
	}
	if pathExists("/tmp") {
		hostConfig.Mounts = append(hostConfig.Mounts, bindIntoLocal("/tmp"))
	}
	if pathExists("/Users") {
		hostConfig.Mounts = append(hostConfig.Mounts, bindIntoLocal("/Users"))
	}

	return &config, &hostConfig, &networkingConfig
}

func (sham *Sham) RunContainer() {
	sham.l.Debug("starting new container")

	config, hostConfig, networkingConfig := sham.BuildContainerConfig()

	stdin, stdout, stderr := term.StdStreams()
	inStream := streams.NewIn(stdin)
	outStream := streams.NewOut(stdout)
	errStream := stderr

	// Telling the Windows daemon the initial size of the tty during start makes
	// a far better user experience rather than relying on subsequent resizes
	// to cause things to catch up.
	if runtime.GOOS == "windows" {
		hostConfig.ConsoleSize[0], hostConfig.ConsoleSize[1] = outStream.GetTtySize()
	}

	ctx, cancelFun := context.WithCancel(sham.ctx)
	defer cancelFun()

	create, err := sham.docker.ContainerCreate(sham.ctx, config, hostConfig, networkingConfig, "")
	if err != nil {
		sham.l.Fatal("Failed to create container: ", err)
	}

	sham.l.Debug("created ", create.ID)

	// sigc := ForwardAllSignals(ctx, dockerCli, createResponse.ID)
	// defer signal.StopCatch(sigc)

	attachOptions := types.ContainerAttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
	}

	attach, errAttach := sham.docker.ContainerAttach(ctx, create.ID, attachOptions)
	if err != nil {
		sham.l.Fatal("Failed to attach container: ", err)
	}
	defer attach.Close()

	errCh := make(chan error, 1)

	go func() {
		errCh <- func() error {
			streamer := hijackedIOStreamer{
				streams:      dockerCli,
				inputStream:  in,
				outputStream: out,
				errorStream:  cerr,
				resp:         resp,
				tty:          config.Tty,
				detachKeys:   options.DetachKeys,
			}

			if errHijack := streamer.stream(ctx); errHijack != nil {
				return errHijack
			}
			return errAttach
		}()
	}()

	cancelFun()
	<-errCh
}

/*
func attachContainer(
	ctx context.Context,
	errCh *chan error,
	config *container.Config,
	containerID string,
) (func(), error) {
	stdout, stderr := dockerCli.Out(), dockerCli.Err()
	var (
		out, cerr io.Writer
		in        io.ReadCloser
	)
	if config.AttachStdin {
		in = dockerCli.In()
	}
	if config.AttachStdout {
		out = stdout
	}
	if config.AttachStderr {
		if config.Tty {
			cerr = stdout
		} else {
			cerr = stderr
		}
	}

	options := types.ContainerAttachOptions{
		Stream:     true,
		Stdin:      config.AttachStdin,
		Stdout:     config.AttachStdout,
		Stderr:     config.AttachStderr,
		DetachKeys: dockerCli.ConfigFile().DetachKeys,
	}

	resp, errAttach := dockerCli.Client().ContainerAttach(ctx, containerID, options)
	if errAttach != nil {
		return nil, errAttach
	}

	ch := make(chan error, 1)
	*errCh = ch

	go func() {
		ch <- func() error {
			streamer := hijackedIOStreamer{
				streams:      dockerCli,
				inputStream:  in,
				outputStream: out,
				errorStream:  cerr,
				resp:         resp,
				tty:          config.Tty,
				detachKeys:   options.DetachKeys,
			}

			if errHijack := streamer.stream(ctx); errHijack != nil {
				return errHijack
			}
			return errAttach
		}()
	}()
	return resp.Close, nil
}
*/

func (sham *Sham) StartContainer() {
	sham.l.Debug("starting new container")

	config, hostConfig, networkingConfig := sham.BuildContainerConfig()
	create, err := sham.docker.ContainerCreate(sham.ctx, config, hostConfig, networkingConfig, "")
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
	return mount.Mount{Type: mount.TypeBind, Source: path, Target: fmt.Sprintf("/host%s", path)}
}
