package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func main() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()

	// options := platform.LoadInitOptionsFromEnv()

	/*
		{
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
			var cmdargs []string
			cmdargs = append(cmdargs, "--home-dir", options.Home)
			cmdargs = append(cmdargs, "--gid", fmt.Sprint(options.Gid))
			cmdargs = append(cmdargs, "--no-create-home")
			cmdargs = append(cmdargs, "--shell", "/bin/bash")
			cmdargs = append(cmdargs, "--uid", fmt.Sprint(options.Uid))
			cmdargs = append(cmdargs, options.Username)

			cmd := exec.Command("useradd", cmdargs...)
			stdoutStderr, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("%s\n", stdoutStderr)
				panic(err)
			}
		}

		err := os.Chown("/run/host-services/ssh-auth.sock", options.Uid, options.Gid)
		if err != nil {
			log.Fatalf("%v", err)
		}

		sudo := []byte(fmt.Sprintf("%s ALL=(ALL) NOPASSWD:ALL\n", options.Username))
		err = ioutil.WriteFile("/etc/sudoers.d/user", sudo, 0644)
		if err != nil {
			panic(err)
		}

		err = ioutil.WriteFile("/toolbox-options.json", []byte(os.Getenv("PLATFORM_OPTIONS")), 0644)
		if err != nil {
			panic(err)
		}
	*/

	// if os.Getenv("TOOLBOX_INIT_WAIT") != "" {
	// log.Printf("Sleeping...")
	// select {}
	// }

	fmt.Println("awaiting TERM signal")
	<-done
	fmt.Println("exiting")
}

func runCmd(cmdline string) {
	log.Printf("# %s", cmdline)
	cmd := exec.Command("/bin/sh", "-c", cmdline)
	stdoutStderr, err := cmd.CombinedOutput()
	fmt.Printf("%s\n", stdoutStderr)
	if err != nil {
		panic(err)
	}
}
