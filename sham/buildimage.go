package sham

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/sirupsen/logrus"
)

func failIfNotFound() bool {
	return true
}

func okIfNotFound() bool {
	return false
}

func (sham *Sham) FindBaseImage(mustExist bool) {
	sham.l.Debug("Finding base image: ", sham.config.Image)

	var options types.ImageListOptions
	options.Filters = filters.NewArgs()
	options.Filters.Add("reference", sham.config.Image)
	images, err := sham.docker.ImageList(sham.ctx, options)
	if err != nil {
		sham.l.Fatal("failed to list images: ", err)
	}

	for i, image := range images {
		sham.l.Debug("Found image ", i, ": ", image.RepoTags, " ", image.ID, " ", image.Created)
	}

	if len(images) > 0 {
		sham.baseImage = &images[0]
		return
	}

	if mustExist {
		sham.l.Fatal("base image must exist but is not found")
	} else {
		sham.l.Info("base image not found")
	}
}

func (sham *Sham) FindShamImage(mustExist bool) {
	sham.l.Debug("finding sham image for: ", sham.config.Name, " ", sham.config.Image)

	var options types.ImageListOptions
	options.Filters = filters.NewArgs()
	options.Filters.Add("reference", fmt.Sprintf("%s-sham:latest", sham.config.Name))
	options.Filters.Add("label", fmt.Sprintf("%s=%s", keyShamName, sham.config.Name))
	options.Filters.Add("label", fmt.Sprintf("%s=%s", keyShamImageRef, sham.config.Image))
	images, err := sham.docker.ImageList(sham.ctx, options)
	if err != nil {
		sham.l.Fatal("Failed to list images: ", err)
	}

	for i, image := range images {
		sham.l.Debug("Found image ", i, ": ", image.RepoTags, " ", image.ID, " ", image.Created)
	}

	if len(images) > 0 {
		sham.shamImage = &images[0]
		return
	}

	if mustExist {
		sham.l.Fatal("sham image must exist but is not found")
	} else {
		sham.l.Info("sham image not found")
	}
}

func (sham *Sham) PullBaseImage() {
	sham.l.Debug("Pulling base image: ", sham.config.Image)

	var options types.ImagePullOptions
	resp, err := sham.docker.ImagePull(sham.ctx, sham.config.Image, options)
	if err != nil {
		sham.l.Fatal("failed to pull image: ", err)
	}
	defer resp.Close()

	scanner := bufio.NewScanner(resp)
	for scanner.Scan() {
		line := scanner.Text()
		sham.l.Debug(line)
	}
	if err := scanner.Err(); err != nil {
		sham.l.Fatal("reading pull output: ", err)
	}

	sham.l.Debug("image pull complete")
}

func (sham *Sham) BuildImage() {
	sham.l.Info("building sham image")

	image := sham.baseImage

	sham.l.Debug("using base image id:", image.ID, " labels:", image.Labels)

	hasher := sha1.New()
	_, err := hasher.Write([]byte(image.ID))
	if err != nil {
		sham.l.Fatal("failed to hash image ID", err)
	}
	encodedID := base64.URLEncoding.EncodeToString(hasher.Sum(nil))[:12]

	initOptions := sham.initOptions.AsString()

	sham.l.Debug("SHAM_INIT_OPTIONS: ", initOptions)

	userID := fmt.Sprintf("%v", sham.initOptions.Uid)

	var buildOptions types.ImageBuildOptions
	buildOptions.RemoteContext = "http://localhost/Dockerfile"
	// buildOptions.Remove = true
	// buildOptions.ForceRemove = true
	// buildOptions.NoCache = true
	buildOptions.Tags = []string{
		fmt.Sprintf("%s-sham:%s", sham.config.Name, encodedID),
		fmt.Sprintf("%s-sham:latest", sham.config.Name),
	}
	buildOptions.Labels = map[string]string{
		keySham:         "",
		keyShamName:     sham.config.Name,
		keyShamImageID:  image.ID,
		keyShamImageRef: sham.config.Image,
	}

	buildOptions.BuildArgs = make(map[string]*string)
	buildOptions.BuildArgs["IMAGE"] = &sham.config.Image
	buildOptions.BuildArgs["USER_ID"] = &userID
	buildOptions.BuildArgs["SHAM_INIT_OPTIONS"] = &initOptions
	buildOptions.BuildArgs["SHAM_BINARY_IMAGE"] = &sham.binaryImageRef

	sham.l.Debug("running build")

	resp, err := sham.docker.ImageBuild(sham.ctx, nil, buildOptions)
	if err != nil {
		sham.l.Fatal("failed to build toolbox: ", err)
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)

	for {
		var jm jsonmessage.JSONMessage
		if err := dec.Decode(&jm); err != nil {
			if err == io.EOF {
				break
			}
			sham.l.Fatal("failed to decode build messages: ", err)
		}

		// output json at trace levels
		if sham.l.Logger.IsLevelEnabled(logrus.TraceLevel) {
			j, err := json.Marshal(jm)
			if err != nil {
				sham.l.Warn("failed to marshal jsonmessage from build: ", err)
			}
			sham.l.Trace("build: ", string(j))
		}

		if jm.Error != nil {
			sham.l.Fatal("sham image build failed: ", jm.Error.Message)
		} else if jm.Stream != "\n" && len(jm.Stream) > 0 {
			sham.l.Info("build stream: ", jm.Stream)
		}
	}

	sham.l.Debug("image build complete")
}
