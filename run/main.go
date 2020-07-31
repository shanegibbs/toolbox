package main

import (
	"log"
	"os"
	"os/exec"
	"syscall"

	platform "github.com/shanegibbs/toolbox/toolbox"
	"gopkg.in/hlandau/service.v1/daemon/setuid"
)

func main() {
	initOptions := platform.LoadInitOptionsFromEnv()
	runOptions := platform.LoadRunOptionsFromEnv()

	platform.ProcessRunOptions(runOptions)

	err := os.Chdir(runOptions.Workdir)
	if err != nil {
		log.Fatalf("Unable to chdir to %s: %v", runOptions.Workdir, err)
	}

	// drop perms
	setuid.Setuid(initOptions.Uid)
	setuid.Setgid(initOptions.Gid)

	if len(runOptions.Args) < 1 {
		log.Fatalf("received no args")
	}

	arg0 := runOptions.Args[0]

	binary, err := exec.LookPath(arg0)
	if err != nil {
		log.Fatalf("unable to find %s", arg0)
	}

	log.Printf("Running in toolbox: %v", runOptions.Args)

	if err := syscall.Exec(binary, runOptions.Args, runOptions.Env); err != nil {
		log.Fatal(err)
	}
}
