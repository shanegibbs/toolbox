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
		sham.FindShamImage()
		if sham.shamImage == nil {
			sham.BuildImage()
		}
		sham.CreateContainer()
	}

	sham.SendCommandToContainer()
}
