package log

import (
	"io"

	"gopkg.in/logfmt.v0"
)

type logfmtLogger struct {
	w io.Writer
}

// NewLogfmtLogger returns a basic logger that encodes keyvals as simple "k=v"
// pairs to the Writer.
func NewLogfmtLogger(w io.Writer) Logger {
	return &logfmtLogger{w}
}

func (l logfmtLogger) Log(keyvals ...interface{}) error {
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
