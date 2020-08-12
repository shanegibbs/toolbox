package sham

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/sirupsen/logrus"
)

func CmdRun(dropPerms func(l *logrus.Entry, initOptions *InitOptions)) {
	sham := New("run")

	initOptions := LoadInitOptionsFromEnv()
	runOptions := LoadRunOptionsFromEnv()

	ProcessRunOptions(runOptions)

	err := os.Chdir(runOptions.Workdir)
	if err != nil {
		sham.l.Fatal("Unable to chdir to ", runOptions.Workdir, ": ", err)
	}

	sham.l.Debug("before uid=", os.Getuid(), "/", os.Geteuid(), " gid=", os.Getgid(), "/", os.Getegid())

	dropPerms(sham.l, initOptions)

	sham.l.Debug("after  uid=", os.Getuid(), "/", os.Geteuid(), " gid=", os.Getgid(), "/", os.Getegid())

	if len(runOptions.Args) < 1 {
		sham.l.Fatal("received no args")
	}

	arg0 := runOptions.Args[0]

	binary, err := exec.LookPath(arg0)
	if err != nil {
		sham.l.Fatal("unable to find ", arg0)
	}

	sham.l.Info("executing: ", runOptions.Args)

	if err := syscall.Exec(binary, runOptions.Args, runOptions.Env); err != nil {
		sham.l.Fatal(err)
	}
}
