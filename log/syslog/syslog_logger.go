package syslog

import (
	gosyslog "log/syslog"
	"io"
	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/log"
	"bytes"
	"sync"
)

// @TODO
// 1. ~Wrap gosyslog.Writer~
// 2. ~Provide dummy implementation for tests~
// 3. ~Test that different levels are sent correctly with default selector~
// 4. ~Test that all levels are sent with a custom exhaustive selector~
// 5. ~Test default value~
// 6. Extract bufLogger - colorLogger


type PrioritySelector func(keyvals ...interface{}) gosyslog.Priority

func NewSyslogLogger(w SyslogWriter, newLogger func(io.Writer) log.Logger, options ...Option) log.Logger {
	l := &syslogLogger{
		w: w,
		newLogger: newLogger,
		prioritySelector: defaultPrioritySelector,
		bufPool: sync.Pool{New: func() interface{} {
			return &loggerBuf{}
		}},
	}

	for _, option := range options {
		option(l)
	}

	return l
}

type syslogLogger struct {
	w                SyslogWriter
	newLogger        func(io.Writer) log.Logger
	prioritySelector PrioritySelector
	bufPool          sync.Pool
}

func (l *syslogLogger) Log(keyvals ...interface{}) error {
	level := l.prioritySelector(keyvals...)

	lb := l.getLoggerBuf()
	defer l.putLoggerBuf(lb)
	if err := lb.logger.Log(keyvals...); err != nil {
		return err
	}
// @TODO - trim newline?
	switch level {
	case gosyslog.LOG_EMERG:
		return l.w.Emerg(lb.buf.String())
	case gosyslog.LOG_ALERT:
		return l.w.Alert(lb.buf.String())
	case gosyslog.LOG_CRIT:
		return l.w.Crit(lb.buf.String())
	case gosyslog.LOG_ERR:
		return l.w.Err(lb.buf.String())
	case gosyslog.LOG_WARNING:
		return l.w.Warning(lb.buf.String())
	case gosyslog.LOG_NOTICE:
		return l.w.Notice(lb.buf.String())
	case gosyslog.LOG_INFO:
		return l.w.Info(lb.buf.String())
	case gosyslog.LOG_DEBUG:
		return l.w.Debug(lb.buf.String())
	default:
		_, err := l.w.Write(lb.buf.Bytes())
		return err
	}
}

type loggerBuf struct {
	buf    *bytes.Buffer
	logger log.Logger
}

func (l *syslogLogger) getLoggerBuf() *loggerBuf {
	lb := l.bufPool.Get().(*loggerBuf)
	if lb.buf == nil {
		lb.buf = &bytes.Buffer{}
		lb.logger = l.newLogger(lb.buf)
	} else {
		lb.buf.Reset()
	}
	return lb
}

func (l *syslogLogger) putLoggerBuf(cb *loggerBuf) {
	l.bufPool.Put(cb)
}

type Option func(*syslogLogger)

func PrioritySelectorOption(selector PrioritySelector) Option {
	return func (l *syslogLogger) { l.prioritySelector = selector }
}

func defaultPrioritySelector(keyvals ...interface{}) gosyslog.Priority {
	for i := 0; i < len(keyvals); i += 2 {
		if keyvals[i] == level.Key() {
			if v, ok := keyvals[i+1].(level.Value); ok {
				switch v {
				case level.DebugValue():
					return gosyslog.LOG_DEBUG
				case level.InfoValue():
					return gosyslog.LOG_INFO
				case level.WarnValue():
					return gosyslog.LOG_WARNING
				case level.ErrorValue():
					return gosyslog.LOG_ERR
				}
			}
		}
	}

	return gosyslog.LOG_INFO
}
