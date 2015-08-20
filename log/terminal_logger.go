package log

import (
	"fmt"
	"io"
	"time"

	"github.com/go-kit/kit/log/term"
	"gopkg.in/logfmt.v0"
)

// For testing, allow to disable TTY detection
var IsTTY = term.IsTty

var _ = Logger((*terminalLogger)(nil))

// NewTerminalLogger returns a log.Logger which prouces nice colored logs,
// but only if the Writer is a tty.
// It is a copy of http://godoc.org/gopkg.in/inconshreveable/log15.v2/#TerminalFormat
//
// Otherwise it will return the alternate logger.
//
// options can be nil, in this case the default options (as returned by NewTerminalOptions) are used.
func NewTerminalLogger(w io.Writer, alternate Logger, options *TerminalOptions) Logger {
	var isTTY bool
	if std, ok := w.(fder); ok {
		isTTY = IsTTY(std.Fd())
	}
	if !isTTY {
		return alternate
	}
	var opts TerminalOptions
	if options == nil {
		opts = *NewTerminalOptions()
	} else {
		opts = *options
	}
	return &terminalLogger{
		w:               w,
		TerminalOptions: opts,
	}
}

type terminalLogger struct {
	w io.Writer
	TerminalOptions
}

func (l terminalLogger) Log(keyvals ...interface{}) error {
	var ts, msg, level string

	for i := 0; i < len(keyvals); i += 2 {
		var found bool
		switch keyvals[i] {
		case l.MsgKey:
			if msg == "" {
				msg = asString(keyvals[i+1])
				found = true
			}
		case l.LevelKey:
			if level == "" {
				level = asString(keyvals[i+1])
				found = true
			}
		case l.TsKey:
			if ts == "" {
				ts = asTimeString(keyvals[i+1], l.TimeFormat)
				found = true
			}
		}
		if found { // delete
			if len(keyvals) == i-2 {
				keyvals = keyvals[:i]
			} else {
				keyvals = append(keyvals[:i], keyvals[i+2:]...)
			}
			i -= 2
		}
	}

	lvl := "INFO"
	// copied from gopkg.in/inconshreveable/log15.v2/format.go
	// ---8<---
	var color = 0
	switch level {
	case l.DebugValue:
		color = 36
		lvl = "DBUG"
	case l.InfoValue:
		color = 32
		lvl = "INFO"
	case l.ErrorValue:
		color = 31
		lvl = "EROR"
	case l.WarnValue:
		color = 33
		lvl = "WARN"
	case l.CritValue:
		color = 35
		lvl = "CRIT"
	}
	if color > 0 {
		fmt.Fprintf(l.w, "\x1b[%dm%s\x1b[0m[%s] %s ", color, lvl, ts, msg)
	} else {
		fmt.Fprintf(l.w, "[%s] [%s] %s ", lvl, ts, msg)
	}
	// --->8---

	// copied from gopkg.in/kit.v0/log/logfmt_logger.go
	// ---8<---
	b, err := logfmt.MarshalKeyvals(keyvals...)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	if _, err := l.w.Write(b); err != nil {
		return err
	}
	return nil
	// --->8---
}

func asString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return x
	case fmt.Stringer:
		return x.String()
	case fmt.Formatter:
		return fmt.Sprint(x)
	default:
		return fmt.Sprintf("%s", x)
	}
}
func asTimeString(v interface{}, timeFormat string) string {
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case time.Time:
		return x.Format(timeFormat)
	default:
		return asString(x)
	}
}

type fder interface {
	Fd() uintptr
}

// TerminalOptions contains the options for a terminalLogger.
type TerminalOptions struct {
	MsgKey     string
	TsKey      string
	LevelKey   string
	TimeFormat string

	DebugValue string
	InfoValue  string
	WarnValue  string
	ErrorValue string
	CritValue  string
}

// NewTerminalOptions returns the default TerminalOptions.
func NewTerminalOptions() *TerminalOptions {
	return &TerminalOptions{
		TimeFormat: time.RFC3339,
		MsgKey:     "msg",
		TsKey:      "ts",
		LevelKey:   "level",

		DebugValue: "debug",
		InfoValue:  "info",
		WarnValue:  "warn",
		ErrorValue: "error",
		CritValue:  "crit",
	}
}
