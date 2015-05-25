package log

import (
	"bytes"
	"fmt"
	"io"
)

type logfmtLogger struct {
	io.Writer
}

// NewLogfmtLogger returns a basic logger that encodes keyvals as simple "k=v"
// pairs to the Writer.
func NewLogfmtLogger(w io.Writer) Logger {
	return &logfmtLogger{w}
}

func (l logfmtLogger) Log(keyvals ...interface{}) error {
	if len(keyvals)%2 == 1 {
		panic("odd number of keyvals")
	}
	buf := &bytes.Buffer{}
	for i := 0; i < len(keyvals); i += 2 {
		if i != 0 {
			if _, err := fmt.Fprint(buf, " "); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(buf, "%s=%v", keyvals[i], keyvals[i+1]); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(l.Writer, buf.String()); err != nil {
		return err
	}
	return nil
}
