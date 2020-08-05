package platform

import (
	"encoding/json"
	"log"
	"os"
	"os/user"
	"strconv"
	"strings"
)

type InitOptions struct {
	Username string
	Home     string
	Uid      int
	Gid      int
}

type RunOptions struct {
	Workdir string
	Args    []string
	Env     []string
}

func (o *InitOptions) AsString() string {
	buf, err := json.Marshal(o)
	if err != nil {
		panic(err)
	}
	return string(buf)
}

func (o *RunOptions) AsString() string {
	buf, err := json.Marshal(o)
	if err != nil {
		panic(err)
	}
	return string(buf)
}

func BuildInitOptions() *InitOptions {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	uid, err := strconv.Atoi(user.Uid)
	gid, err := strconv.Atoi(user.Gid)

	options := &InitOptions{
		Username: user.Username,
		Home:     user.HomeDir,
		Uid:      uid,
		Gid:      gid,
	}

	return options
}

func BuildRunOptions() *RunOptions {
	workdir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	options := &RunOptions{
		Workdir: workdir,
		Args:    os.Args,
		Env:     os.Environ(),
	}

	return options
}

const initOptionsEnvKey = "SHAM_INIT_OPTIONS"

func LoadInitOptionsFromEnv() *InitOptions {
	var options InitOptions
	env, exists := os.LookupEnv(initOptionsEnvKey)
	if !exists {
		log.Fatalf("%s not found", initOptionsEnvKey)
	}
	// log.Printf("%s=%s", initOptionsEnvKey, env)
	err := json.Unmarshal([]byte(env), &options)
	if err != nil {
		log.Fatalf("Failed to parse init options: %v\n%s", err, env)
	}
	return &options
}

func LoadRunOptionsFromEnv() *RunOptions {
	var options RunOptions
	env, exists := os.LookupEnv("TOOLBOX_RUN_OPTIONS")
	if !exists {
		log.Fatalf("TOOLBOX_RUN_OPTIONS not found")
	}
	// log.Printf("TOOLBOX_RUN_OPTIONS=%s", env)
	err := json.Unmarshal([]byte(env), &options)
	if err != nil {
		log.Fatalf("Failed to parse run options: %v\n%s", err, env)
	}
	return &options
}

func ProcessRunOptions(options *RunOptions) {
	options.Env = filter(options.Env, func(e string) bool {
		if strings.HasPrefix(e, "SSH_AUTH_SOCK=") {
			return false
		}

		if strings.HasPrefix(e, "TMPDIR=") {
			return false
		}

		if strings.HasPrefix(e, "PATH=") {
			return false
		}

		return true
	})

	options.Env = append(options.Env, "SSH_AUTH_SOCK=/run/host-services/ssh-auth.sock")

	if strings.Contains(options.Args[0], "stub") || strings.Contains(options.Args[0], "toolbox") || strings.HasSuffix(options.Args[0], "exe/main") {
		options.Args = options.Args[1:]
	}

	if len(options.Args) == 0 {
		options.Args = []string{"bash", "-il"}
	}

	options.Env = append(options.Env, "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin")
}

func filter(arr []string, cond func(string) bool) []string {
	result := []string{}
	for i := range arr {
		if cond(arr[i]) {
			result = append(result, arr[i])
		}
	}
	return result
}
