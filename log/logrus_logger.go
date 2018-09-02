package log

import (
	"fmt"
	"io"

	"github.com/sirupsen/logrus"
)

type logrusLogger struct {
	*logrus.Logger
}

// NewLogrusLogger returns a logger that logs the keyvals to Writer
// with a time stamp at the logrus.InfoLevel
func NewLogrusLogger(w io.Writer) Logger {
	newLogger := logrus.New()
	newLogger.Out = w
	newLogger.Formatter = &logrus.TextFormatter{TimestampFormat: "02-01-2006 15:04:05", FullTimestamp: true}
	return &logrusLogger{newLogger}
}

func (l logrusLogger) Log(keyvals ...interface{}) error {
	if len(keyvals)%2 == 0 {
		fields := logrus.Fields{}
		for i := 0; i < len(keyvals); i += 2 {
			fields[fmt.Sprint(keyvals[i])] = keyvals[i+1]
		}
		l.WithFields(fields).Info()
	} else {
		l.Info(keyvals)
	}
	return nil
}
