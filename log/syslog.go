// +build linux,!appengine darwin freebsd openbsd

package log

import (
	"io"
	"log/syslog"

	"github.com/go-kit/kit/log/level"
)

type syslogWriter struct {
	*syslog.Writer
	selector func(keyvals ...interface{}) syslog.Priority
}

type SyslogAdapterOption interface {
	Apply(*syslogWriter)
}

func NewSyslogWriter(w *syslog.Writer, options ...SyslogAdapterOption) io.Writer {
	writer := &syslogWriter{
		syslog.Writer: w,
		selector:      defaultSyslogSelector,
	}

	for _, option := range options {
		option.Apply(writer)
	}

	return *writer
}

type syslogWriterAdapter struct {
	f func(string) error
}

func (a *syslogWriterAdapter) Write(b []byte) (int, error) {
	return len(b), a.f(string(b))
}

func (w syslogWriter) GetSpecializedWriter(keyvals ...interface{}) io.Writer {
	priority := w.selector(keyvals...)

	switch priority {
	case syslog.LOG_DEBUG:
		return syslogWriterAdapter{f: w.Debug}
	case syslog.LOG_INFO:
		return syslogWriterAdapter{f: w.Info}
	case syslog.LOG_WARN:
		return syslogWriterAdapter{f: w.Warn}
	case syslog.LOG_ERR:
		return syslogWriterAdapter{f: w.Error}
	}

	return w
}

func defaultSyslogSelector(keyvals ...interface{}) syslog.Priority {
	for i := 1; i < len(keyvals); i += 2 {
		if v, ok := keyvals[i].(*Value); ok {
			switch v {
			case level.DebugValue():
				return syslog.LOG_DEBUG
			case level.InfoValue():
				return syslog.LOG_INFO
			case level.WarnValue():
				return syslog.LOG_WARN
			case level.ErrorValue():
				return syslog.LOG_ERR
			}
		}
	}

	return syslog.LOG_INFO
}

// SyslogPrioritySelector is an option that specifies the syslog priority
// selector.
type SyslogPrioritySelector struct {
	PrioritySelector func(keyvals ...interface{}) syslog.Priority
}

func (o SyslogPrioritySelector) Apply(w *syslogWriter) {
	w.selector = o.PrioritySelector
}
