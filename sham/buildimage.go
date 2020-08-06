package sham

import (
	"bufio"
	"fmt"

	"github.com/docker/docker/api/types"
	log "github.com/sirupsen/logrus"
)

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
