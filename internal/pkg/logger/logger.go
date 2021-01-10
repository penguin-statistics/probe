package logger

import (
	"github.com/sirupsen/logrus"
)

// New creates a new logrus logger
func New(module string) *logrus.Entry {
	l := logrus.New()
	l.SetFormatter(&logrus.TextFormatter{TimestampFormat: "2006-01-02 15:04:05.000", FullTimestamp: true})
	l.SetLevel(logrus.DebugLevel)
	return l.WithField("module", module)
}
