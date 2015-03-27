package log

import (
	"fmt"
	"io"
)

type kvLogger struct {
	io.Writer
	fields []Field
}

// NewKVLogger returns a Logger that prefixes fields as space-separated "k=v"
// pairs on every log line.
func NewKVLogger(w io.Writer) Logger {
	return &kvLogger{
		Writer: w,
	}
}

func (l *kvLogger) With(fields ...Field) Logger {
	return &kvLogger{
		Writer: l.Writer,
		fields: append(l.fields, fields...),
	}
}

func (l *kvLogger) Log(s string) error {
	for _, f := range l.fields {
		if _, err := fmt.Fprintf(l.Writer, "%s=%v ", f.Key, f.Value); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(l.Writer, s); err != nil {
		return err
	}
	if s[len(s)-1] != '\n' {
		if _, err := fmt.Fprintf(l.Writer, "\n"); err != nil {
			return err
		}
	}
	return nil
}
