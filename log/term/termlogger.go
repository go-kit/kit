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

// NewTerminalLogger returns a log.Logger which formats its output as
// [TIME] [LEVEL] MESAGE key=value key=value ...
// (see github.com/inconshreveable/log15).
//
// The output is NOT colored, for that wrap it in a ColorLogger.
//
// The options can be nil, in this case the default options
// (as returned by NewLogOpts) are used.
//
// Example usage:
//	logger := log.NewLogfmtLogger(os.Stderr)
//	if term.IsTerminal(os.Stderr) {
//		logger = term.NewTerminalLogger(os.Stderr, nil)
//	}
func NewTerminalLogger(w io.Writer, options *LogOpts) log.Logger {
	var opts LogOpts
	if options == nil {
		opts = *NewLogOpts()
	} else {
		opts = *options
	}
	return &terminalLogger{w: w, LogOpts: opts}
}

// Color returns the color to be used printing this log record.
// This function uses the level (the value where the key is opts.LevelKey and the value is one of opts.DebugValue ... opts.CritValue) to determine the color.
//
// For customization, the easiest is to copy this little code and change the colors.
func (opts LogOpts) Color(keyvals ...interface{}) FgBgColor {
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

// NewLogger returns a new TerminaLogger - this is a convenience function
// for calling opts.NewLogger(w) instead of NewTerminalLogger(w, opts).
//
// This can be used for example to ease wrapping with ColorLoger:
//    NewColorLogger(w, opts.NewLogger, opts.Color)
func (opts LogOpts) NewLogger(w io.Writer) log.Logger {
	return &terminalLogger{w: w, LogOpts: opts}
}

// NewColorLogger is a convenience function for returning a log.Logger
// as NewColorLogger(w, opts.NewLogger, opts.Color) does.
//
// The options can be nil, in this case the default options
// (as returned by NewLogOpts) are used.
//
// Example usage:
//	logger := log.NewLogfmtLogger(os.Stderr)
//	if term.IsTerminal(os.Stderr) {
//		logger = term.NewLogOpts().NewColorLogger(os.Stderr)
//	}
func (opts LogOpts) NewColorLogger(w io.Writer) log.Logger {
	return NewColorLogger(w, opts.NewLogger, opts.Color)
}

type terminalLogger struct {
	w io.Writer
	LogOpts

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

	// copied from github.com/go-kit/kit/log/logfmt_logger.go
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

// LogOpts contains the options for a terminalLogger.
type LogOpts struct {
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

// NewLogOpts returns the default LogOpts.
func NewLogOpts() *LogOpts {
	return &LogOpts{
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
