package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// JSONFieldBehavior dictates how fields are encoded in a JSON logger.
type JSONFieldBehavior int

const (
	// PrefixedFields encodes each field as its own JSON object, and prefixes
	// each log event with those objects, separated by spaces.
	PrefixedFields JSONFieldBehavior = iota

	// MixedFields encodes each field into the same JSON object as the log
	// event. Logged events take precedence.
	MixedFields
)

// NewJSONLogger returns a Logger that encodes log events as JSON objects.
// Logged events are expected to be valid JSON.
func NewJSONLogger(w io.Writer, fieldBehavior JSONFieldBehavior) Logger {
	return &jsonLogger{w, fieldBehavior, []Field{}}
}

type jsonLogger struct {
	io.Writer
	JSONFieldBehavior
	fields []Field
}

func (l *jsonLogger) With(f Field) Logger {
	return &jsonLogger{
		Writer:            l.Writer,
		JSONFieldBehavior: l.JSONFieldBehavior,
		fields:            append(l.fields, f),
	}
}

func (l *jsonLogger) Logf(format string, args ...interface{}) error {
	var buf bytes.Buffer
	if _, err := fmt.Fprintf(&buf, format, args...); err != nil {
		return err
	}

	m := map[string]interface{}{}
	if err := json.NewDecoder(&buf).Decode(&m); err != nil {
		return err
	}

	buf.Reset()
	switch l.JSONFieldBehavior {
	case PrefixedFields:
		if err := EncodeMany(&buf, JSON, l.fields); err != nil {
			return err
		}
		if err := json.NewEncoder(&buf).Encode(m); err != nil {
			return err
		}

	case MixedFields:
		final := map[string]interface{}{}
		for _, f := range l.fields {
			final[f.Key] = f.Value // fields first, so that...
		}
		for k, v := range m {
			final[k] = v // ...logged data takes precedence
		}
		if err := json.NewEncoder(&buf).Encode(final); err != nil {
			return err
		}
	}

	_, err := buf.WriteTo(l.Writer)
	return err
}
