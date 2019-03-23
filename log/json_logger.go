package log

import (
	"encoding"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

type jsonLogger struct {
	io.Writer
}

// NewJSONLogger returns a Logger that encodes keyvals to the Writer as a
// single JSON object. Each log event produces no more than one call to
// w.Write. The passed Writer must be safe for concurrent use by multiple
// goroutines if the returned Logger will be used concurrently.
func NewJSONLogger(w io.Writer) Logger {
	return &jsonLogger{w}
}

func (l *jsonLogger) Log(keyvals ...interface{}) error {
	n := (len(keyvals) + 1) / 2 // +1 to handle case when len is odd
	m := make(map[string]interface{}, n)
	for i := 0; i < len(keyvals); i += 2 {
		k := keyvals[i]
		var v interface{} = ErrMissingValue
		if i+1 < len(keyvals) {
			v = keyvals[i+1]
		}
		merge(m, k, v)
	}
	return json.NewEncoder(l.Writer).Encode(m)
}

func merge(dst map[string]interface{}, k, v interface{}) {
	var key string
	switch x := k.(type) {
	case string:
		key = x
	case fmt.Stringer:
		key = safeString(x)
	default:
		key = fmt.Sprint(x)
	}

	// We want json.Marshaler and encoding.TextMarshaller to take priority over
	// err.Error() and v.String(). But json.Marshall (called later) does that by
	// default so we force a no-op if it's one of those 2 case.
	switch x := v.(type) {
	case json.Marshaler:
	case encoding.TextMarshaler:
	case error:
		v = safeError(x)
	case fmt.Stringer:
		v = safeString(x)
	}

	dst[key] = v
}

func safeString(str fmt.Stringer) (s string) {
	defer func() {
		if panicVal := recover(); panicVal != nil {
			if v := reflect.ValueOf(str); v.Kind() == reflect.Ptr && v.IsNil() {
				s = "NULL"
			} else {
				panic(panicVal)
			}
		}
	}()
	s = str.String()
	return
}

func safeError(err error) (s interface{}) {
	defer func() {
		if panicVal := recover(); panicVal != nil {
			if v := reflect.ValueOf(err); v.Kind() == reflect.Ptr && v.IsNil() {
				s = nil
			} else {
				panic(panicVal)
			}
		}
	}()
	s = err.Error()
	return
}

type jsonArrayWriter struct {
	events []json.RawMessage
}

// NewJSONArrayWriter returns an io.Writer that accumulates log events
// in memory, so that you can emit them all as sub-events within a single
// outer event, by passing the writer itself as a log value. This can
// help to achieve the modern observability
// recommendation to emit one very wide event per request.
//
// The accumulation happens by appending to a slice, so if you
// need to guarantee goroutine-safe logging, wrap it with NewSyncWriter.
func NewJSONArrayWriter() io.Writer {
	return &jsonArrayWriter{[]json.RawMessage{}}
}

func (w *jsonArrayWriter) Write(p []byte) (int, error) {
	w.events = append(w.events, json.RawMessage(append([]byte{}, p...)))
	return len(p), nil
}

func (w *jsonArrayWriter) MarshalJSON() ([]byte, error) {
	return json.Marshal(w.events)
}
