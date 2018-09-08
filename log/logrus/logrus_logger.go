package log

import (
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/sirupsen/logrus"
)

type logrusLogger struct {
	*logrus.Logger
}

// NewLogrusLogger takes a *logrus.Logger and returns
//a logger that stisfies the go-kit log.Logger interface
func NewLogrusLogger(logger *logrus.Logger) log.Logger {
	return &logrusLogger{logger}
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
