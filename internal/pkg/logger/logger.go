package logger

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// New creates a new logrus logger
func New(module string) *logrus.Entry {
	l := logrus.New()
	l.SetFormatter(&logrus.TextFormatter{TimestampFormat: "2006-01-02 15:04:05.000", FullTimestamp: true})
	if viper.GetBool("app.debug") {
		l.SetLevel(logrus.TraceLevel)
		l.SetReportCaller(true)
	}
	return l.WithField("module", module)
}
