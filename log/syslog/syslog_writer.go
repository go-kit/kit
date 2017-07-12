package syslog

import (
	gosyslog "log/syslog"
)

type SyslogWriter interface {
	Write([]byte) (int, error)
	Close() error
	Emerg(string) error
	Alert(string) error
	Crit(string) error
	Err(string) error
	Warning(string) error
	Notice(string) error
	Info(string) error
	Debug(string) error
}

type syslogWriter struct {
	gosyslog.Writer
}