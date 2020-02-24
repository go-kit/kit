// Package logrus provides an adapter to the
// go-kit log.Logger interface.
package logrus

import (
	"errors"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/sirupsen/logrus"
)

type Logger struct {
	logrus.FieldLogger
	logrus.Level
}

type Option func(*Logger)

var errMissingValue = errors.New("(MISSING)")

// NewLogger returns a go-kit log.Logger that sends log events to a Logrus logger.
func NewLogger(logger logrus.FieldLogger, options ...Option) log.Logger {
	l := &Logger{
		FieldLogger: logger,
		Level:       logrus.InfoLevel,
	}

	for _, optFunc := range options {
		optFunc(l)
	}

	return l
}

// WithLevel configures a logrus logger to set specific log
// level to log messages with
func WithLevel(level logrus.Level) Option {
	return func(c *Logger) {
		c.Level = level
	}
}

func (l Logger) Log(keyvals ...interface{}) error {
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
