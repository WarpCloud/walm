package util

import (
	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func init() {
	Log = logrus.New()
	//logrus.SetFormatter(new(logrus.TextFormatter))
}
