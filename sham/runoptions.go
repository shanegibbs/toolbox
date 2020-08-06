package sham

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

const runOptionsEnvKey = "SHAM_RUN_OPTIONS"

type RunOptions struct {
	Workdir string
	Args    []string
	Env     []string
}

func (o *RunOptions) AsString() string {
	buf, err := json.Marshal(o)
	if err != nil {
		panic(err)
	}
	return string(buf)
}

func (sham *Sham) BuildRunOptions() {
	workdir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	args := os.Args

	path := fmt.Sprintf("/Users/shane.gibbs/.sham/shams/%s/", sham.config.Name)
	if strings.HasPrefix(args[0], path) {
		sham.l.Info("detected sham ", path)
		args[0] = args[0][len(path):]
		sham.l.Info("set arg0 ", args[0])
	}

	options := &RunOptions{
		Workdir: workdir,
		Args:    args,
		Env:     os.Environ(),
	}

	sham.runOptions = options
}

func LoadRunOptionsFromEnv() *RunOptions {
	var options RunOptions
	env, exists := os.LookupEnv(runOptionsEnvKey)
	if !exists {
		log.Fatal("env not found: ", runOptionsEnvKey)
	}
	log.Trace(runOptionsEnvKey, "=", env)
	err := json.Unmarshal([]byte(env), &options)
	if err != nil {
		log.Fatal("Failed to parse run options: ", err, "\n", env)
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

	if strings.Contains(options.Args[0], "stub") || strings.Contains(options.Args[0], "sham") || strings.HasSuffix(options.Args[0], "exe/main") {
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
