package log

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type Logger interface {
	New(keyvals ...interface{}) Logger
	SetHandler(h Handler)
	Log(keyvals ...interface{})
}

type logger struct {
	ctx []interface{}
	h   Handler
}

func New(keyvals ...interface{}) Logger {
	return &logger{
		ctx: expandKeyVals(keyvals),
	}
}

func (l *logger) New(keyvals ...interface{}) Logger {
	var kv []interface{}
	kv = append(kv, l.ctx...)
	kv = append(kv, expandKeyVals(keyvals)...)

	return &logger{
		ctx: kv,
		h:   l.h,
	}
}

func (l *logger) SetHandler(h Handler) {
	l.h = h
}

func (l *logger) Log(keyvals ...interface{}) {
	var kv []interface{}
	kv = append(kv, l.ctx...)
	// kv = append(keyvals, CalllerKey, stack.Caller(1))
	kv = append(kv, expandKeyVals(keyvals)...)
	l.h.Handle(kv...)
}

func expandKeyVals(keyvals []interface{}) []interface{} {
	kvCount := len(keyvals)
	for _, kv := range keyvals {
		if _, ok := kv.(KeyVal); ok {
			kvCount++
		}
	}
	if kvCount == len(keyvals) {
		return keyvals
	}
	exp := make([]interface{}, 0, kvCount)
	for _, kv := range keyvals {
		switch kv := kv.(type) {
		case KeyVal:
			exp = append(exp, kv.Key())
			exp = append(exp, kv.Value())
		default:
			exp = append(exp, kv)
		}
	}
	return exp
}

type KeyVal interface {
	Key() interface{}
	Value() interface{}
}

type key int

const (
	TimeKey key = iota
	LvlKey
	CalllerKey
)

func Now() KeyVal {
	return logtime(time.Now())
}

type logtime time.Time

func (t logtime) Key() interface{}   { return TimeKey }
func (t logtime) Value() interface{} { return time.Time(t) }

type Lvl int

const (
	Crit Lvl = iota
	Error
	Warn
	Info
	Debug
)

func (l Lvl) Key() interface{}   { return LvlKey }
func (l Lvl) Value() interface{} { return l }

func (l Lvl) MarshalText() (text []byte, err error) {
	switch l {
	case Crit:
		return []byte("crit"), nil
	case Error:
		return []byte("eror"), nil
	case Warn:
		return []byte("warn"), nil
	case Info:
		return []byte("info"), nil
	case Debug:
		return []byte("dbug"), nil
	}
	panic("unexpected level value")
}

type Handler interface {
	Handle(keyvals ...interface{}) error
}

type HandlerFunc func(keyvals ...interface{}) error

func (f HandlerFunc) Handle(keyvals ...interface{}) error {
	return f(keyvals...)
}

type Encoder interface {
	Encode(keyvals ...interface{}) ([]byte, error)
}

func AddTimestamp(h Handler) Handler {
	return HandlerFunc(func(keyvals ...interface{}) error {
		kvs := append(keyvals, nil, nil)
		copy(kvs[2:], kvs[:len(kvs)-2])
		kvs[0] = TimeKey
		kvs[1] = time.Now()
		return h.Handle(kvs...)
	})
}

func ReplaceKeys(h Handler, keypairs ...interface{}) Handler {
	m := make(map[interface{}]interface{}, len(keypairs)/2)
	for i := 0; i < len(keypairs); i += 2 {
		m[keypairs[i]] = keypairs[i+1]
	}
	return HandlerFunc(func(keyvals ...interface{}) error {
		for i := 0; i < len(keyvals); i += 2 {
			if nk, ok := m[keyvals[i]]; ok {
				keyvals[i] = nk
			}
		}
		return h.Handle(keyvals...)
	})
}

func Writer(w io.Writer, enc Encoder) Handler {
	return HandlerFunc(func(keyvals ...interface{}) error {
		b, err := enc.Encode(keyvals...)
		if err != nil {
			return err
		}
		_, err = w.Write(b)
		return err
	})
}

type EncoderFunc func(keyvals ...interface{}) ([]byte, error)

func (f EncoderFunc) Encode(keyvals ...interface{}) ([]byte, error) {
	return f(keyvals...)
}

func JsonEncoder() Encoder {
	return EncoderFunc(func(keyvals ...interface{}) ([]byte, error) {
		m := make(map[string]interface{}, len(keyvals)/2)
		for i := 0; i < len(keyvals); i += 2 {
			m[fmt.Sprint(keyvals[i])] = keyvals[i+1]
		}
		b, err := json.Marshal(m)
		if err != nil {
			return nil, err
		}
		return b, nil
	})
}

// func Failover(alternates ...Handler) Handler  { return nil }
// func LvlFilter(maxLvl Lvl, h Handler) Handler { return nil }
// func Multiple(hs ...Handler) Handler          { return nil }

// func Discard() Handler                        { return nil }
// func Encode(enc Encoder, w io.Writer) Handler { return nil }
