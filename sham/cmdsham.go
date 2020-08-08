package sham

func CmdSham() {
	// if os.Args {
	// }
	sham := New()
	sham.SetupLogging("sham")
	sham.LoadConfig()
	sham.CreateDockerClient()
	sham.BuildInitOptions()
	sham.BuildRunOptions()

	sham.FindShamContainer()
	if sham.shamContainer != nil {
		sham.SendCommandToContainer()
	}

	sham.FindShamImage(okIfNotFound())
	if sham.shamImage != nil {
		sham.CreateContainer()
		sham.InstallShams()
		sham.SendCommandToContainer()
	}

	sham.FindBaseImage(okIfNotFound())
	if sham.baseImage != nil {
		sham.BuildImage()
		sham.FindShamImage(failIfNotFound())
		sham.CreateContainer()
		sham.InstallShams()
		sham.SendCommandToContainer()
	}

	sham.PullBaseImage()
	sham.FindBaseImage(failIfNotFound())
	sham.BuildImage()
	sham.FindShamImage(failIfNotFound())
	sham.CreateContainer()
	sham.InstallShams()
	sham.SendCommandToContainer()
}
