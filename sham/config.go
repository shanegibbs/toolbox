package sham

import (
	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Config struct {
	l           *logrus.Entry
	Name        string
	Image       string
	Args        []string
	DefaultArgs []string
	Shams       []string
}

func (this *Config) Load() {
	this.LoadFrom(os.Args)
}

func (this *Config) LoadFrom(args []string) {
	this.DefaultArgs = []string{"bash", "-il"}

	if args[1] == "run" && len(args) > 2 {
		this.Name = args[2]
		this.Image = args[2]
		this.Args = os.Args[3:]

		if len(this.Args) == 0 {
			this.Args = this.DefaultArgs
		}

		return
	}

	bodyBytes, err := ioutil.ReadFile("sham.yaml")
	if err != nil {
		this.l.Fatal("Failed to read sham.yaml: ", err)
	}
	body := string(bodyBytes)

	this.l.Debug("config:\n", body)

	var config Config
	err = yaml.Unmarshal(bodyBytes, &config)
	if err != nil {
		this.l.Fatal("failed to parse config: ", err)
	}
}
