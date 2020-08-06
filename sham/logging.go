package sham

import (
	"os"

	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

func (sham *Sham) SetupLogging(prefix string) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
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
