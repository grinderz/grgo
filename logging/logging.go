package logging

import (
	"github.com/sirupsen/logrus"
)

var (
	Log *logrus.Logger
)

func init() {
	Log = logrus.New()
}

func Configure(logLevel string, logCaller bool) {
	lvl, err := logrus.ParseLevel(logLevel)
	if err != nil {
		lvl = logrus.DebugLevel
	}
	Log.SetLevel(lvl)
	Log.SetReportCaller(logCaller)
}
