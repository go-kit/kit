package term

import (
	"bytes"
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

	maxColor
)

var resetColorBytes = []byte("\x1b[0m")
var fgColorBytes [][]byte
var bgColorBytes [][]byte

func init() {
	for color := NoColor; color < maxColor; color++ {
		fgColorBytes = append(fgColorBytes, []byte(fmt.Sprintf("\x1b[%dm", 30+color)))
		bgColorBytes = append(bgColorBytes, []byte(fmt.Sprintf("\x1b[%dm", 40+color)))
	}
}

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
func NewColorLogger(w io.Writer, newLogger func(io.Writer) log.Logger, color func(keyvals ...interface{}) FgBgColor) log.Logger {
	if color == nil {
		panic("color func nil")
	}
	return &colorLogger{
		w:             w,
		newLogger:     newLogger,
		color:         color,
		bufPool:       sync.Pool{New: func() interface{} { return &loggerBuf{} }},
		noColorLogger: newLogger(w),
	}
}

type colorLogger struct {
	w             io.Writer
	newLogger     func(io.Writer) log.Logger
	color         func(keyvals ...interface{}) FgBgColor
	bufPool       sync.Pool
	noColorLogger log.Logger
}

func (l *colorLogger) Log(keyvals ...interface{}) error {
	color := l.color(keyvals...)
	if color.IsZero() {
		return l.noColorLogger.Log(keyvals...)
	}

	lb := l.getLoggerBuf()
	defer l.putLoggerBuf(lb)
	if color.Fg != NoColor {
		lb.buf.Write(fgColorBytes[color.Fg])
	}
	if color.Bg != NoColor {
		lb.buf.Write(bgColorBytes[color.Bg])
	}
	err := lb.logger.Log(keyvals...)
	if err != nil {
		return err
	}
	if color.Fg != NoColor || color.Bg != NoColor {
		lb.buf.Write(resetColorBytes)
	}
	_, err = io.Copy(l.w, lb.buf)
	return err
}

type loggerBuf struct {
	buf    *bytes.Buffer
	logger log.Logger
}

func (l *colorLogger) getLoggerBuf() *loggerBuf {
	lb := l.bufPool.Get().(*loggerBuf)
	if lb.buf == nil {
		lb.buf = &bytes.Buffer{}
		lb.logger = l.newLogger(lb.buf)
	} else {
		lb.buf.Reset()
	}
	return lb
}

func (l *colorLogger) putLoggerBuf(cb *loggerBuf) {
	l.bufPool.Put(cb)
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
