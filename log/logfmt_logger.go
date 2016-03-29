package log

import (
	"io"

	"github.com/go-logfmt/logfmt"
)

type logfmtLogger struct {
	w io.Writer
}

// NewLogfmtLogger returns a logger that encodes keyvals to the Writer in
// logfmt format. The passed Writer must be safe for concurrent use by
// multiple goroutines if the returned Logger will be used concurrently.
func NewLogfmtLogger(w io.Writer) Logger {
	return &logfmtLogger{w}
}

func (l logfmtLogger) Log(keyvals ...interface{}) error {
	// The Logger interface requires implementations to be safe for concurrent
	// use by multiple goroutines. For this implementation that means making
	// only one call to l.w.Write() for each call to Log. We first collect all
	// of the bytes into b, and then call l.w.Write(b).
	b, err := logfmt.MarshalKeyvals(keyvals...)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	if _, err := l.w.Write(b); err != nil {
		return err
	}
	return nil
}
