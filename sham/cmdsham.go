package sham

func CmdSham() {
	sham := New()
	sham.SetupLogging("sham")
	sham.CreateDockerClient()
	sham.BuildInitOptions()
	sham.BuildRunOptions()

	id := sham.FindExistingContainerID()

	if id == nil {
		sham.BuildImage()
		id = sham.CreateContainer()
	}

	sham.SendCommandToContainer()
}
