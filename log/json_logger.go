package log

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sync"
)

type jsonEncoder struct {
	*json.Encoder
	buf bytes.Buffer
}

func (l *jsonEncoder) Reset() {
	l.buf.Reset()
}

var jsonEncoderPool = sync.Pool{
	New: func() interface{} {
		var enc jsonEncoder
		enc.Encoder = json.NewEncoder(&enc.buf)
		return &enc
	},
}

type jsonLogger struct {
	w io.Writer
}

// NewJSONLogger returns a Logger that encodes keyvals to the Writer as a
// single JSON object.
func NewJSONLogger(w io.Writer) Logger {
	return &jsonLogger{w}
}

func (l *jsonLogger) Log(keyvals ...interface{}) error {
	enc := jsonEncoderPool.Get().(*jsonEncoder)
	enc.Reset()
	defer jsonEncoderPool.Put(enc)

	enc.buf.WriteByte('{')

	for i := 0; i < len(keyvals); i += 2 {
		if i > 0 {
			enc.buf.WriteByte(',')
		}
		k := keyvals[i]
		var key string
		switch x := k.(type) {
		case string:
			key = x
		case fmt.Stringer:
			key, _ = safeString(x)
		default:
			key = fmt.Sprint(x)
		}
		writeQuotedString(&enc.buf, key)

		enc.buf.WriteByte(':')

		var v interface{} = ErrMissingValue
		if i+1 < len(keyvals) {
			v = keyvals[i+1]
		}

		if err := writeValue(enc, v); err != nil {
			return err
		}
	}

	enc.buf.WriteString("}\n")

	// The Logger interface requires implementations to be safe for concurrent
	// use by multiple goroutines. For this implementation that means making
	// only one call to l.w.Write() for each call to Log.
	_, err := l.w.Write(enc.buf.Bytes())
	return err
}

var null = []byte("NULL")

func writeValue(enc *jsonEncoder, value interface{}) error {
	w := &enc.buf

	switch v := value.(type) {
	case nil:
		_, err := w.Write(null)
		return err
	case string:
		return writeStringValue(w, v, true)
	case json.Marshaler:
		vb, err := safeJSONMarshal(v)
		if err != nil {
			return err
		}
		if vb == nil {
			vb = null
		}
		_, err = w.Write(vb)
		return err
	case encoding.TextMarshaler:
		vb, err := safeMarshal(v)
		if err != nil {
			return err
		}
		if vb == nil {
			vb = null
		}
		_, err = writeQuotedBytes(w, vb)
		return err
	case error:
		se, ok := safeError(v)
		return writeStringValue(w, se, ok)
	case fmt.Stringer:
		ss, ok := safeString(v)
		return writeStringValue(w, ss, ok)
	default:
		if err := enc.Encode(value); err != nil {
			return err
		}

		// Remove trailing newline added by encoder
		enc.buf.Truncate(enc.buf.Len() - 1)
		return nil
	}
}

func writeStringValue(w io.Writer, value string, ok bool) error {
	var err error
	if !ok && value == "null" {
		_, err = io.WriteString(w, `null`)
	} else {
		_, err = writeQuotedString(w, value)
	}
	return err
}

func safeError(err error) (s string, ok bool) {
	defer func() {
		if panicVal := recover(); panicVal != nil {
			if v := reflect.ValueOf(err); v.Kind() == reflect.Ptr && v.IsNil() {
				s, ok = "null", false
			} else {
				panic(panicVal)
			}
		}
	}()
	s, ok = err.Error(), true
	return
}

func safeString(str fmt.Stringer) (s string, ok bool) {
	defer func() {
		if panicVal := recover(); panicVal != nil {
			if v := reflect.ValueOf(str); v.Kind() == reflect.Ptr && v.IsNil() {
				s, ok = "NULL", false
			} else {
				panic(panicVal)
			}
		}
	}()
	s, ok = str.String(), true
	return
}

// MarshalerError represents an error encountered while marshaling a value.
type MarshalerError struct {
	Type reflect.Type
	Err  error
}

func (e *MarshalerError) Error() string {
	return "error marshaling value of type " + e.Type.String() + ": " + e.Err.Error()
}

func safeJSONMarshal(jm json.Marshaler) (b []byte, err error) {
	defer func() {
		if panicVal := recover(); panicVal != nil {
			if v := reflect.ValueOf(jm); v.Kind() == reflect.Ptr && v.IsNil() {
				b, err = nil, nil
			} else {
				panic(panicVal)
			}
		}
	}()
	b, err = jm.MarshalJSON()
	if err != nil {
		return nil, &MarshalerError{
			Type: reflect.TypeOf(jm),
			Err:  err,
		}
	}
	return
}

func safeMarshal(tm encoding.TextMarshaler) (b []byte, err error) {
	defer func() {
		if panicVal := recover(); panicVal != nil {
			if v := reflect.ValueOf(tm); v.Kind() == reflect.Ptr && v.IsNil() {
				b, err = nil, nil
			} else {
				panic(panicVal)
			}
		}
	}()
	b, err = tm.MarshalText()
	if err != nil {
		return nil, &MarshalerError{
			Type: reflect.TypeOf(tm),
			Err:  err,
		}
	}
	return
}
