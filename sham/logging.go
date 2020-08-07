package sham

import (
	"os"

	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

func (sham *Sham) SetupLogging(prefix string) {
	log := logrus.New()

	{
		val, exist := os.LookupEnv("SHAM_LOG")
		if exist && val == "fatal" {
			log.SetLevel(logrus.FatalLevel)
		} else if exist && val == "error" {
			log.SetLevel(logrus.ErrorLevel)
		} else if exist && val == "info" {
			log.SetLevel(logrus.InfoLevel)
		} else if exist && val == "debug" {
			log.SetLevel(logrus.DebugLevel)
		} else if exist && val == "trace" {
			log.SetLevel(logrus.TraceLevel)
		} else {
			log.SetLevel(logrus.InfoLevel)
		}
	}

	// logrus.SetFormatter(&logrus.TextFormatter{
	// 	DisableColors: false,
	// 	ForceColors:   true,
	// 	FullTimestamp: true,
	// 	PadLevelText:  true,
	// })
	// log.SetReportCaller(true)

	log.Formatter = &prefixed.TextFormatter{
		ForceColors:      true,
		DisableColors:    false,
		DisableUppercase: true,
		FullTimestamp:    true,
	}

	l := log.WithField("prefix", prefix)
	l.Info("args: ", os.Args)

	sham.l = l
}
