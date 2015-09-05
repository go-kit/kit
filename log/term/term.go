// Package term provides tools for logging to a terminal.
package term

import (
	"io"

	"github.com/go-kit/kit/log"
)

// NewLogger returns a Logger that takes advantage of terminal features if
// possible. Log events are formatted by the Logger returned by newLogger. If
// w is a terminal each log event is colored according to the color function.
func NewLogger(w io.Writer, newLogger func(io.Writer) log.Logger, color func(keyvals ...interface{}) FgBgColor) log.Logger {
	fw, ok := w.(FdWriter)
	if !ok || !IsTerminal(fw.Fd()) {
		return newLogger(w)
	}
	return NewColorLogger(NewColorWriter(fw), newLogger, color)
}

// An FdWriter is a Writer that has a file descriptor.
type FdWriter interface {
	io.Writer
	Fd() uintptr
}
