package log

import (
	"encoding/json"
	"fmt"
	"io"
)

type jsonLogger struct {
	io.Writer
	key     string
	keyvals []interface{}
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

func (l *jsonLogger) With(keyvals ...interface{}) Logger {
	if len(keyvals)%2 == 1 {
		panic("odd number of keyvals")
	}
	return &jsonLogger{
		Writer:  l.Writer,
		key:     l.key,
		keyvals: append(l.keyvals, keyvals...),
	}
}

func (l *jsonLogger) Log(message string, keyvals ...interface{}) error {
	if len(keyvals)%2 == 1 {
		panic("odd number of keyvals")
	}
	m := make(map[string]interface{}, len(l.keyvals)+len(keyvals)+1)
	for i := 0; i < len(l.keyvals); i += 2 {
		merge(m, l.keyvals[i], l.keyvals[i+1])
	}
	for i := 0; i < len(keyvals); i += 2 {
		merge(m, keyvals[i], keyvals[i+1])
	}
	m[l.key] = message
	return json.NewEncoder(l.Writer).Encode(m)
}

func merge(dst map[string]interface{}, k, v interface{}) map[string]interface{} {
	var key string
	switch x := k.(type) {
	case string:
		key = x
	case fmt.Stringer:
		key = x.String()
	default:
		key = fmt.Sprintf("%v", x)
	}
	if x, ok := v.(error); ok {
		v = x.Error()
	}
	dst[key] = v
	return dst
}
