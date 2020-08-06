package sham

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"syscall"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const keySham = "com.gibbsdevops.sham"
const keyShamName = "com.gibbsdevops.sham.name"
const keyShamImageID = "com.gibbsdevops.sham.image.id"
const keyShamImageRef = "com.gibbsdevops.sham.image.ref"

type Sham struct {
	ctx           context.Context
	l             *logrus.Entry
	initOptions   *InitOptions
	runOptions    *RunOptions
	docker        *client.Client
	config        *Config
	baseImage     *types.ImageSummary
	shamImage     *types.ImageSummary
	shamContainer *types.Container
}

type Config struct {
	Name  string
	Image string
}

func New() *Sham {
	l := logrus.WithField("prefix", "foo")
	l.Trace("Start. Args: ", os.Args)
	return &Sham{
		ctx: context.Background(),
		l:   l,
	}
}

func (sham *Sham) LoadConfig() {
	bodyBytes, err := ioutil.ReadFile("sham.yaml")
	if err != nil {
		sham.l.Fatal("Failed to read sham.yaml", err)
	}
	body := string(bodyBytes)

	sham.l.Debug("config:\n", body)

	var config Config
	err = yaml.Unmarshal(bodyBytes, &config)
	if err != nil {
		sham.l.Fatal("failed to parse config: ", err)
	}

	sham.config = &config
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
	args = append(args, sham.shamContainer.ID)
	args = append(args, "/sham-run")

	os.Setenv("SHAM_INIT_OPTIONS", sham.initOptions.AsString())
	os.Setenv("SHAM_RUN_OPTIONS", sham.runOptions.AsString())

	// hand proc off to docker
	sham.l.Debug("handing off to docker")
	if err := syscall.Exec("/usr/local/bin/docker", args, os.Environ()); err != nil {
		sham.l.Fatal(err)
	}
}
