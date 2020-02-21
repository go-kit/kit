// Package logrus provides an adapter to the
// go-kit log.Logger interface.
package logrus

import (
	"errors"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/sirupsen/logrus"
)

type logrusLogger struct {
	logrus.FieldLogger
	logrus.Level
}

var errMissingValue = errors.New("(MISSING)")

// NewLogrusLogger returns a go-kit log.Logger that sends log events to a Logrus logger.
func NewLogrusLogger(logger logrus.FieldLogger) log.Logger {
	return &logrusLogger{logger, logrus.InfoLevel}
}

// NewLogrusLoggerWithLevel returns a go-kit log.Logger that sends log events to a Logrus logger, which
// will be logged with provided error level
func NewLogrusLoggerWithLevel(logger logrus.FieldLogger, level logrus.Level) log.Logger {
	return &logrusLogger{logger, level}
}

func (l logrusLogger) Log(keyvals ...interface{}) error {
	fields := logrus.Fields{}
	for i := 0; i < len(keyvals); i += 2 {
		if i+1 < len(keyvals) {
			fields[fmt.Sprint(keyvals[i])] = keyvals[i+1]
		} else {
			fields[fmt.Sprint(keyvals[i])] = errMissingValue
		}
	}

	switch l.Level {
	case logrus.InfoLevel:
		l.WithFields(fields).Info()
	case logrus.ErrorLevel:
		l.WithFields(fields).Error()
	case logrus.DebugLevel:
		l.WithFields(fields).Debug()
	case logrus.WarnLevel:
		l.WithFields(fields).Warn()
	default:
		l.WithFields(fields).Print()
	}

	return nil
}
