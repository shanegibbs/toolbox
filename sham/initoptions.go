package sham

import (
	"encoding/json"
	"os"
	"os/user"
	"strconv"

	log "github.com/sirupsen/logrus"
)

type InitOptions struct {
	Username string
	Home     string
	Uid      int
	Gid      int
}

func (o *InitOptions) AsString() string {
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

const initOptionsEnvKey = "SHAM_INIT_OPTIONS"

func LoadInitOptionsFromEnv() *InitOptions {
	var options InitOptions
	env, exists := os.LookupEnv(initOptionsEnvKey)
	if !exists {
		log.Fatal("env not found: ", initOptionsEnvKey)
	}
	log.Trace(initOptionsEnvKey, "=", env)
	err := json.Unmarshal([]byte(env), &options)
	if err != nil {
		log.Fatal("Failed to parse init options: ", err, "\n", env)
	}
	return &options
}
