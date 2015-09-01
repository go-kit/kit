package log

import (
	"io"
	"sync"

	"gopkg.in/logfmt.v0"
)

type logfmtLogger struct {
	w  io.Writer
	mu sync.RWMutex
}

// NewLogfmtLogger returns a logger that encodes keyvals to the Writer in
// logfmt format. The passed Writer must be safe for concurrent use by
// multiple goroutines if the returned Logger will be used concurrently.
func NewLogfmtLogger(w io.Writer) Logger {
	return &logfmtLogger{w: w}
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
	l.mu.RLock()
	w := l.w
	l.mu.RUnlock()
	if _, err := w.Write(b); err != nil {
		return err
	}
	return nil
}

func (l *logfmtLogger) Hijack(f func(io.Writer) io.Writer) {
	l.mu.Lock()
	l.w = f(l.w)
	l.mu.Unlock()
}
