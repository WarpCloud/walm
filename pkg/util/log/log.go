package util

import (
	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func init() {
	Log = logrus.New()
	Log.Formatter = &logrus.TextFormatter{FullTimestamp: true, DisableColors: true}
}
