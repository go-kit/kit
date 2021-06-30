// Package zerolog provides an adapter to the
// go-kit log.Logger interface.
package zerolog

import (
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/rs/zerolog"
)

type zerologLogger struct {
	Logger *zerolog.Logger
}

func (l zerologLogger) Log(kv ...interface{}) error {
	fields := make(map[string]interface{})

	for i := 0; i < len(kv); i += 2 {
		if i+1 < len(kv) {
			fields[fmt.Sprint(kv[i])] = kv[i+1]
		} else {
			fields[fmt.Sprint(kv[i])] = "(MISSING)"
		}
	}

	var e *zerolog.Event

	switch fmt.Sprint(fields["level"]) {
	case "debug":
		e = l.Logger.Debug()
	case "info":
		e = l.Logger.Info()
	case "warn":
		e = l.Logger.Warn()
	case "error":
		e = l.Logger.Error()
	default:
		e = l.Logger.Debug()
	}

	msg, exists := fields["msg"]
	if !exists {
		msg = ""
	}

	delete(fields, "level") // level key will be added by zerolog
	delete(fields, "msg")
	e.Fields(fields).Msg(fmt.Sprint(msg))

	return nil
}

// NewZerologLogger returns a Go kit log.Logger that sends
// log events to a zerolog.Logger.
func NewZerologLogger(logger *zerolog.Logger) log.Logger {
	zlog := zerologLogger{
		Logger: logger,
	}

	return zlog
}
