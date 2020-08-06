package sham

func CmdSham() {
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

	if sham.shamContainer == nil {
		sham.FindShamImage(okIfNotFound())

		if sham.shamImage == nil {
			sham.FindBaseImage()
			if sham.baseImage == nil {
				sham.PullBaseImage()
				sham.FindBaseImage()
				if sham.baseImage == nil {
					sham.l.Fatal("Unable to find image after pulling: ", sham.config.Image)
				}
			}
			sham.BuildImage()
			sham.FindShamImage(failIfNotFound())
		}
		sham.CreateContainer()
	}

	sham.SendCommandToContainer()
}
