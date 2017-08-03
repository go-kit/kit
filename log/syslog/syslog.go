// +build linux,!appengine darwin freebsd openbsd

package syslog

import (
	"io"
	gosyslog "log/syslog"

	"github.com/go-kit/kit/log/level"
)

type syslogWriter struct {
	*gosyslog.Writer
	selector func(keyvals ...interface{}) gosyslog.Priority
}

type SyslogAdapterOption interface {
	Apply(*syslogWriter)
}

func NewSyslogWriter(w *gosyslog.Writer, options ...SyslogAdapterOption) io.Writer {
	writer := &syslogWriter{
		Writer:   w,
		selector: defaultSyslogSelector,
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
	case gosyslog.LOG_DEBUG:
		return &syslogWriterAdapter{f: w.Debug}
	case gosyslog.LOG_INFO:
		return &syslogWriterAdapter{f: w.Info}
	case gosyslog.LOG_WARNING:
		return &syslogWriterAdapter{f: w.Warning}
	case gosyslog.LOG_ERR:
		return &syslogWriterAdapter{f: w.Err}
	}

	return w
}

func defaultSyslogSelector(keyvals ...interface{}) gosyslog.Priority {
	for i := 1; i < len(keyvals); i += 2 {
		if v, ok := keyvals[i].(level.Value); ok {
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

	return gosyslog.LOG_INFO
}

// SyslogPrioritySelector is an option that specifies the syslog priority
// selector.
type SyslogPrioritySelector struct {
	PrioritySelector func(keyvals ...interface{}) gosyslog.Priority
}

func (o SyslogPrioritySelector) Apply(w *syslogWriter) {
	w.selector = o.PrioritySelector
}
