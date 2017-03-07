package level

import (
	"io"
	"log/syslog"

	"github.com/go-kit/kit/log"
)

type syslogWriterAdapter struct {
	write func(string) error
}

func (w syslogWriterAdapter) Write(b []byte) (int, error) {
	return len(b), w.write(string(b))
}

// NewSyslogLogger returns a Logger which writes to syslog.
func NewSyslogLogger(w *syslog.Writer, newLogger func(io.Writer) log.Logger) log.Logger {
	logger := &syslogLogger{
		w:       w,
		loggers: make(map[Value]log.Logger),
	}

	logger.loggers[debugValue] = newLogger(syslogWriterAdapter{write: w.Debug})
	logger.loggers[infoValue] = newLogger(syslogWriterAdapter{write: w.Info})
	logger.loggers[warnValue] = newLogger(syslogWriterAdapter{write: w.Warning})
	logger.loggers[errorValue] = newLogger(syslogWriterAdapter{write: w.Err})

	return logger
}

type syslogLogger struct {
	w       *syslog.Writer
	loggers map[Value]log.Logger
}

func (l *syslogLogger) Log(keyvals ...interface{}) error {
	level := infoValue

	for i := 1; i < len(keyvals); i += 2 {
		if v, ok := keyvals[i].(*levelValue); ok {
			level = v
			break
		}
	}

	return l.loggers[level].Log(keyvals...)
}
