package log

import (
	"fmt"
	"io"
)

type prefixLogger struct {
	io.Writer
	keyvals []interface{}
}

// NewPrefixLogger returns a basic logger that encodes all keyvals as simple
// "k=v" pairs prefixed on each log line.
func NewPrefixLogger(w io.Writer) Logger {
	return &prefixLogger{
		Writer: w,
	}
}

func (l *prefixLogger) With(keyvals ...interface{}) Logger {
	if len(keyvals)%2 == 1 {
		panic("odd number of keyvals")
	}
	return &prefixLogger{
		Writer:  l.Writer,
		keyvals: append(l.keyvals, keyvals...),
	}
}

func (l *prefixLogger) Log(message string, keyvals ...interface{}) error {
	if len(keyvals)%2 == 1 {
		panic("odd number of keyvals")
	}
	if err := encodeMany(l.Writer, l.keyvals...); err != nil {
		return err
	}
	if err := encodeMany(l.Writer, keyvals...); err != nil {
		return err
	}
	if _, err := fmt.Fprint(l.Writer, message); err != nil {
		return err
	}
	if message[len(message)-1] != '\n' {
		if _, err := fmt.Fprintln(l.Writer); err != nil {
			return err
		}
	}
	return nil
}

func encodeMany(w io.Writer, keyvals ...interface{}) error {
	if len(keyvals)%2 == 1 {
		panic("odd number of keyvals")
	}
	for i := 0; i < len(keyvals); i += 2 {
		_, err := fmt.Fprintf(w, "%s=%v ", keyvals[i], keyvals[i+1])
		if err != nil {
			return err
		}
	}
	return nil
}
