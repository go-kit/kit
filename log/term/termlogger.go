package term

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"gopkg.in/logfmt.v0"
)

var _ = log.Logger((*terminalLogger)(nil))

// NewTerminalLogger returns a log.Logger which prouces nice colored logs.
// It uses ColorLogger for coloring the whole line based on "level".
//
// The options can be nil, in this case the default options
// (as returned by NewTermLogOpts) are used.
//
// Example usage:
//	logger := log.NewLogfmtLogger(os.Stderr)
//	if term.IsTerminal(os.Stderr) {
//		logger = term.NewTerminalLogger(os.Stderr, nil)
//	}
func NewTerminalLogger(w io.Writer, options *TermLogOpts) log.Logger {
	var opts TermLogOpts
	if options == nil {
		opts = *NewTermLogOpts()
	} else {
		opts = *options
	}
	return NewColorLogger(
		w,
		func(w io.Writer) log.Logger {
			return &terminalLogger{w: w, TermLogOpts: opts}
		},
		opts.Color,
	)
}

func (opts TermLogOpts) Color(keyvals ...interface{}) FgBgColor {
	for i := 0; i < len(keyvals); i += 2 {
		if asString(keyvals[i]) == opts.LevelKey {
			switch asString(keyvals[i+1]) {
			case opts.DebugValue:
				return FgBgColor{Fg: Green}
			case opts.InfoValue:
			case opts.WarnValue:
				return FgBgColor{Fg: Yellow}
			case opts.ErrorValue:
				return FgBgColor{Fg: Red}
			case opts.CritValue:
				return FgBgColor{Fg: Default, Bg: Red}
			}
			return FgBgColor{}
		}
	}
	return FgBgColor{}
}

type terminalLogger struct {
	w io.Writer
	TermLogOpts

	buf bytes.Buffer
	mu  sync.Mutex
}

func (l *terminalLogger) Log(keyvals ...interface{}) error {
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
	switch level {
	case l.DebugValue:
		lvl = "DBUG"
	case l.InfoValue:
		lvl = "INFO"
	case l.ErrorValue:
		lvl = "EROR"
	case l.WarnValue:
		lvl = "WARN"
	case l.CritValue:
		lvl = "CRIT"
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.buf.Reset()
	fmt.Fprintf(&l.buf, "[%s] [%s] %s ", lvl, ts, msg)

	// copied from gopkg.in/kit.v0/log/logfmt_logger.go
	// ---8<---
	b, err := logfmt.MarshalKeyvals(keyvals...)
	if err != nil {
		return err
	}
	l.buf.Write(b)
	l.buf.WriteByte('\n')
	if _, err := l.w.Write(l.buf.Bytes()); err != nil {
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

// TermLogOpts contains the options for a terminalLogger.
type TermLogOpts struct {
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

// NewTermLogOpts returns the default TermLogOpts.
func NewTermLogOpts() *TermLogOpts {
	return &TermLogOpts{
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
