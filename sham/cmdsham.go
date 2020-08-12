package sham

func CmdSham() {
	sham := New("sham")
	sham.config.Load()
	sham.CreateDockerClient()
	sham.BuildInitOptions()
	sham.BuildRunOptions()

	sham.FindShamContainer()
	if sham.shamContainer != nil {
		sham.SendCommandToContainer()
	}

	sham.FindShamImage(okIfNotFound())
	if sham.shamImage != nil {
		sham.StartContainer()
		sham.InstallShams()
		sham.SendCommandToContainer()
	}

	sham.FindBaseImage(okIfNotFound())
	if sham.baseImage != nil {
		sham.BuildImage()
		sham.FindShamImage(failIfNotFound())
		sham.StartContainer()
		sham.InstallShams()
		sham.SendCommandToContainer()
	}

	sham.PullBaseImage()
	sham.FindBaseImage(failIfNotFound())
	sham.BuildImage()
	sham.FindShamImage(failIfNotFound())
	sham.StartContainer()
	sham.InstallShams()
	sham.SendCommandToContainer()
}
