package log

import (
	"encoding/json"

	"io"
)

type jsonLogger struct {
	io.Writer
	key    string
	fields []Field
}

// NewJSONLogger returns a Logger that marshals each log line as a JSON
// object. Because fields are keys in a JSON object, they must be unique, and
// last-writer-wins. The actual log message is placed under the "msg" key. To
// change that, use the NewJSONLoggerWithKey constructor.
func NewJSONLogger(w io.Writer) Logger {
	return NewJSONLoggerWithKey(w, "msg")
}

// NewJSONLoggerWithKey is the same as NewJSONLogger but allows the user to
// specify the key under which the actual log message is placed in the JSON
// object.
func NewJSONLoggerWithKey(w io.Writer, messageKey string) Logger {
	return &jsonLogger{
		Writer: w,
		key:    messageKey,
	}
}

func (l *jsonLogger) With(fields ...Field) Logger {
	return &jsonLogger{
		Writer: l.Writer,
		key:    l.key,
		fields: append(l.fields, fields...),
	}
}

func (l *jsonLogger) Log(s string) error {
	m := make(map[string]interface{}, len(l.fields)+1)
	for _, f := range l.fields {
		m[f.Key] = f.Value
	}
	m[l.key] = s
	return json.NewEncoder(l.Writer).Encode(m)
}
