package main

import (
	sham "github.com/shanegibbs/toolbox/sham"
	"github.com/sirupsen/logrus"
)

import (
	//#include <unistd.h>
	//#include <errno.h>
	"C"
)

func main() {
	sham.CmdRun(func(l *logrus.Entry, initOptions *sham.InitOptions) {
		l.Debug("setting uid=", initOptions.Uid, " gui=", initOptions.Gid)
		cerr, errno := C.setgid(C.__gid_t(initOptions.Gid))
		if cerr != 0 {
			l.Fatal("Unable to set GID due to error: ", errno)
		}
		cerr, errno = C.setuid(C.__uid_t(initOptions.Uid))
		if cerr != 0 {
			l.Fatal("Unable to set UID due to error: ", errno)
		}
	})
}
