package log

import (
	"fmt"
	"github.com/najeira/ltsv"
	"io"
	"reflect"
)

type ltsvLogger struct {
	w io.Writer
}

// NewLTSVLogger returns a Logger that encodes keyvals to the Writer as
// Labeled Tab-separated Values format.
func NewLTSVLogger(w io.Writer) Logger {
	return &ltsvLogger{
		w: w,
	}
}

func (l *ltsvLogger) Log(keyvals ...interface{}) error {
	n := (len(keyvals) + 1) / 2 // +1 to handle case when len is odd
	m := make(map[string]string, n)
	for i := 0; i < len(keyvals); i += 2 {
		k := keyvals[i]
		var v interface{} = ErrMissingValue
		if i+1 < len(keyvals) {
			v = keyvals[i+1]
		}
		mergeStringMap(m, k, v)
	}

	w := ltsv.NewWriter(l.w)
	if err := w.Write(m); err != nil {
		return err
	}
	w.Flush()
	return nil
}

func mergeStringMap(dst map[string]string, k, v interface{}) {
	var key string
	switch x := k.(type) {
	case string:
		key = x
	case fmt.Stringer:
		key = safeString(x)
	default:
		key = fmt.Sprint(x)
	}

	var val string
	switch x := v.(type) {
	case string:
		val = x
	case error:
		val = safeErrorString(x)
	case fmt.Stringer:
		val = safeString(x)
	default:
		val = fmt.Sprint(x)
	}

	dst[key] = val
}

func safeErrorString(err error) (s string) {
	defer func() {
		if panicVal := recover(); panicVal != nil {
			if v := reflect.ValueOf(err); v.Kind() == reflect.Ptr && v.IsNil() {
				s = ""
			} else {
				panic(panicVal)
			}
		}
	}()
	s = err.Error()
	return
}
