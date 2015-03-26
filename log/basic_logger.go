package log

import (
	"fmt"
	"io"
)

// NewBasicLogger is the default, simple implementation of a Logger. Fields
// are encoded and prefixed to every log entry.
func NewBasicLogger(w io.Writer, e FieldEncoder) Logger {
	return &basicLogger{w, e, []Field{}}
}

type basicLogger struct {
	io.Writer
	FieldEncoder
	fields []Field
}

func (l *basicLogger) With(f Field) Logger {
	return &basicLogger{
		Writer:       l.Writer,
		FieldEncoder: l.FieldEncoder,
		fields:       append(l.fields, f),
	}
}

func (l *basicLogger) Logf(format string, args ...interface{}) error {
	var err error
	if err = EncodeMany(l.Writer, l.FieldEncoder, l.fields); err != nil {
		return err // TODO: should we continue, best-effort?
	}
	if _, err = fmt.Fprintf(l.Writer, format, args...); err != nil {
		return err
	}
	if len(format) > 0 && format[len(format)-1] != '\n' {
		_, err = fmt.Fprintf(l.Writer, "\n")
	}
	return err
}
