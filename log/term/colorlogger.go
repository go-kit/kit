package term

import (
	"fmt"
	"io"
	"sync"

	"github.com/go-kit/kit/log"
)

// Color is the abstract color, the zero value is the Default.
type Color uint8

const (
	NoColor = Color(iota)
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
	Default
)

type FgBgColor struct {
	Fg, Bg Color
}

func (c FgBgColor) IsZero() bool {
	return c.Fg == NoColor && c.Bg == NoColor
}

var _ = log.Logger((*colorLogger)(nil))

// NewColorLogger returns a log.Logger which produces nice colored logs.
// It colors whole records based on the FgBgColor returned by the color function.
//
// For example for such a function, see LevelColor.
//
// Example for coloring errors with red:
//
//	logger := term.NewColorLogger(log.NewLogfmtLogger(os.Stdout),
//		func(keyvals ...interface) term.FgBgColor {
//			for i := 1; i < len(keyvals); i += 2 {
//				if keyvals[i] != nil {
//					continue
//				}
//				if _, ok := keyvals[i].(error) {
//					return term.FgBgColor{Fg: term.White, Bg: term.Red}
//				}
//			}
//			return term.FgBgColor{}
//		})
//
//	logger.Log("c", "c is uncolored value", "err", nil)
//	logger.Log("c", "c is colored 'cause err colors it", "err", errors.New("coloring error"))
func NewColorLogger(logger log.Logger, color func(keyvals ...interface{}) FgBgColor) log.Logger {
	// FIXME(tgulacsi): I don't understand why can't I use log.Hijacker here...
	hj, ok := logger.(interface {
		Hijack(func(io.Writer) io.Writer)
	})
	if !ok {
		return logger
	}
	cl := &colorLogger{
		logger: logger,
	}
	cl.color = color // otherwise, no coloring is possible!
	hj.Hijack(func(w io.Writer) io.Writer {
		cl.w = w
		return cl
	})
	return cl
}

type colorLogger struct {
	logger log.Logger
	color  func(keyvals ...interface{}) FgBgColor
	w      io.Writer
	mu     sync.RWMutex // protects w

	actColor   FgBgColor
	actColorMu sync.RWMutex // protects actColor
}

func (l *colorLogger) Log(keyvals ...interface{}) error {
	l.actColorMu.Lock() // Unlock is in Write!
	if l.color != nil {
		l.actColor = l.color(keyvals...)
	}
	l.actColorMu.Unlock()
	return l.logger.Log(keyvals...)
}

func (l *colorLogger) Write(p []byte) (int, error) {
	l.actColorMu.RLock()
	color := l.actColor
	l.actColorMu.RUnlock()
	if color.IsZero() {
		return l.w.Write(p)
	}
	var n int
	if color.Fg != NoColor {
		m, err := fmt.Fprintf(l.w, "\x1b[%dm", 30+color.Fg)
		if err != nil {
			return m, err
		}
		n += m
	}
	if color.Bg != NoColor {
		m, err := fmt.Fprintf(l.w, "\x1b[%dm", 40+color.Bg)
		if err != nil {
			return n + m, err
		}
		n += m
	}
	m, err := l.w.Write(p)
	if err != nil {
		return n + m, err
	}
	n += m
	l.mu.RLock()
	w := l.w
	l.mu.RUnlock()
	m, err = w.Write([]byte("\x1b[0m"))
	return n + m, err
}

func (l *colorLogger) Hijack(f func(io.Writer) io.Writer) {
	l.mu.Lock()
	l.w = f(l.w)
	l.mu.Unlock()
}

func asString(v interface{}) string {
	switch x := v.(type) {
	case string:
		return x
	case fmt.Stringer:
		return x.String()
	case fmt.Formatter:
		return fmt.Sprint(x)
	default:
		return fmt.Sprintf("%v", x)
	}
}

// LevelColor returns the color for the record based on the value of the "level" key, if exists.
func LevelColor(keyvals ...interface{}) FgBgColor {
	for i := 0; i < len(keyvals); i += 2 {
		if asString(keyvals[i]) != "level" {
			continue
		}
		switch asString(keyvals[i+1]) {
		case "debug":
			return FgBgColor{Fg: Green}
		case "info":
			return FgBgColor{Fg: White}
		case "warn":
			return FgBgColor{Fg: Yellow}
		case "error":
			return FgBgColor{Fg: Red}
		case "crit":
			return FgBgColor{Fg: Default, Bg: Red}
		default:
			return FgBgColor{}
		}
	}
	return FgBgColor{}
}
