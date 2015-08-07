package log

import (
	"encoding/json"
	"fmt"
	"io"
)

type jsonLogger struct {
	io.Writer
}

// NewJSONLogger returns a Logger that encodes keyvals to the Writer as a
// single JSON object.
func NewJSONLogger(w io.Writer) Logger {
	return &jsonLogger{w}
}

func (l *jsonLogger) Log(keyvals ...interface{}) error {
	n := (len(keyvals) + 1) / 2 // +1 to handle case when len is odd
	m := make(map[string]interface{}, n)
	for i := 0; i < len(keyvals); i += 2 {
		k := keyvals[i]
		v := ErrMissingValue
		if i+1 < len(keyvals) {
			v = keyvals[i+1]
		}
		merge(m, k, v)
	}
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
