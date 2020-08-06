package sham

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func CmdSetup() {
	options := LoadInitOptionsFromEnv()

	{
		log.Printf("Creating group")

		var cmdargs []string
		cmdargs = append(cmdargs, "--gid", fmt.Sprint(options.Gid))
		cmdargs = append(cmdargs, options.Username)

		cmd := exec.Command("groupadd", cmdargs...)
		stdoutStderr, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("%s\n", stdoutStderr)
			panic(err)
		}
	}

	{
		log.Printf("Creating user")

		var cmdargs []string
		cmdargs = append(cmdargs, "--home-dir", options.Home)
		cmdargs = append(cmdargs, "--gid", fmt.Sprint(options.Gid))
		cmdargs = append(cmdargs, "--no-create-home")
		cmdargs = append(cmdargs, "--shell", "/bin/bash")
		cmdargs = append(cmdargs, "--uid", fmt.Sprint(options.Uid))
		cmdargs = append(cmdargs, "--no-log-init")
		cmdargs = append(cmdargs, options.Username)

		cmd := exec.Command("useradd", cmdargs...)
		stdoutStderr, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("%s\n", stdoutStderr)
			panic(err)
		}
	}

	// err := os.Chown("/run/host-services/ssh-auth.sock", options.Uid, options.Gid)
	// if err != nil {
	// 	log.Fatalf("%v", err)
	// }

	log.Printf("Creating sudoers.d")
	err := os.MkdirAll("/etc/sudoers.d", os.ModeDir)
	if err != nil {
		panic(err)
	}

	// sudo := []byte(fmt.Sprintf("%s ALL=(ALL) NOPASSWD:ALL\n", options.Username))
	// err = ioutil.WriteFile("/etc/sudoers.d/user", sudo, 0644)
	// if err != nil {
	// 	panic(err)
	// }

	// err = ioutil.WriteFile("/toolbox-options.json", []byte(os.Getenv("sham_OPTIONS")), 0644)
	// if err != nil {
	// 	panic(err)
	// }

	log.Printf("--sham-setup-complete--")
}
