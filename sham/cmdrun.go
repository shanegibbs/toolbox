package sham

import (
	"os"
	"os/exec"
	"syscall"

	log "github.com/sirupsen/logrus"
	"gopkg.in/hlandau/service.v1/daemon/setuid"
)

func CmdRun() {
	sham := New()
	sham.SetupLogging("run")

	initOptions := LoadInitOptionsFromEnv()
	runOptions := LoadRunOptionsFromEnv()

	ProcessRunOptions(runOptions)

	err := os.Chdir(runOptions.Workdir)
	if err != nil {
		log.Fatal("Unable to chdir to ", runOptions.Workdir, ": ", err)
	}

	// drop perms
	setuid.Setuid(initOptions.Uid)
	setuid.Setgid(initOptions.Gid)

	if len(runOptions.Args) < 1 {
		log.Fatal("received no args")
	}

	arg0 := runOptions.Args[0]

	binary, err := exec.LookPath(arg0)
	if err != nil {
		log.Fatal("unable to find ", arg0)
	}

	sham.l.Info("executing: ", runOptions.Args)

	if err := syscall.Exec(binary, runOptions.Args, runOptions.Env); err != nil {
		log.Fatal(err)
	}
}
