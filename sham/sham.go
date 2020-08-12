package sham

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

const keySham = "com.gibbsdevops.sham"
const keyShamName = "com.gibbsdevops.sham.name"
const keyShamImageID = "com.gibbsdevops.sham.image.id"
const keyShamImageRef = "com.gibbsdevops.sham.image.ref"

type Sham struct {
	ctx            context.Context
	l              *logrus.Entry
	initOptions    *InitOptions
	runOptions     *RunOptions
	docker         *client.Client
	config         *Config
	binaryImageRef string
	baseImage      *types.ImageSummary
	shamImage      *types.ImageSummary
	shamContainer  *types.Container
}

func New(name string) *Sham {
	sham := &Sham{
		ctx:            context.Background(),
		binaryImageRef: "sham",
	}

	sham.SetupLogging(name)
	sham.l.Trace("Start. Args: ", os.Args)

	var config Config
	config.l = sham.l
	sham.config = &config

	return sham
}

func (sham *Sham) InstallShams() {
	sham.l.Trace("installing shams ", sham.config.Shams)

	path := fmt.Sprintf("/Users/shane.gibbs/.sham/shams/%s", sham.config.Name)

	// TODO check if exists first, error if can't delte
	os.RemoveAll(path)

	os.MkdirAll(path, 0755)

	for _, s := range sham.config.Shams {
		shamPath := fmt.Sprintf("%s/%s", path, s)

		sham.l.Info("installing sham ", shamPath)
		os.Symlink("/Users/shane.gibbs/bin/sham", shamPath)
	}
}

func (sham *Sham) BuildInitOptions() {
	sham.initOptions = BuildInitOptions()
}

func (sham *Sham) CreateDockerClient() {
	docker, err := client.NewEnvClient()
	if err != nil {
		sham.l.Fatalf("failed to create docker client: ", err)
	}
	sham.l.Debug("created docker client")
	sham.docker = docker
}

func (sham *Sham) FindShamContainer() {
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
		if container.State != "running" {
			continue
		}
		for _, name := range container.Names {
			sham.l.Debug("Found sham container ID: ", container.ID, " with name: ", name, " ", container.State, "/", container.Status)
		}
		sham.shamContainer = &container
		return
	}

	sham.l.Info("sham container not found")
}

func (sham *Sham) inTerminal() bool {
	fi, _ := os.Stdin.Stat()
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		sham.l.Debug("stdin is from pipe")
		return false
	} else {
		sham.l.Debug("stdin is from terminal")
		return true
	}
}

func (sham *Sham) SendCommandToContainer() {
	var args []string
	args = append(args, "docker", "exec")
	args = append(args, "-i")
	if sham.inTerminal() {
		args = append(args, "-t")
	}
	args = append(args, "--env", "SHAM_INIT_OPTIONS")
	args = append(args, "--env", "SHAM_RUN_OPTIONS")
	args = append(args, "--env", "SHAM_LOG")
	args = append(args, sham.shamContainer.ID)
	args = append(args, "/sham-run")

	os.Setenv("SHAM_INIT_OPTIONS", sham.initOptions.AsString())
	os.Setenv("SHAM_RUN_OPTIONS", sham.runOptions.AsString())

	docker, err := exec.LookPath("docker")
	if err != nil {
		sham.l.Fatal("docker not found: ", err)
	}

	// hand proc off to docker
	sham.l.Info("handing off to container ", sham.shamContainer.Names[0])
	if err := syscall.Exec(docker, args, os.Environ()); err != nil {
		sham.l.Fatal("hand off failed: ", err)
	}
}
