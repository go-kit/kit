package log

import (
	"time"

	"gopkg.in/stack.v1"
)

// BindValues returns a slice with all value elements (odd indexes) that
// implement Valuer replaced with the result of calling their Value method. If
// no value elements implement Valuer, the original slice is returned.
func BindValues(keyvals ...interface{}) []interface{} {
	if !containsValuer(keyvals) {
		return keyvals
	}

	bound := make([]interface{}, len(keyvals))
	copy(bound, keyvals)
	for i := 1; i < len(bound); i += 2 {
		if v, ok := bound[i].(Valuer); ok {
			bound[i] = v.Value()
		}
	}

	return bound
}

func containsValuer(keyvals []interface{}) bool {
	for i := 1; i < len(keyvals); i += 2 {
		if _, ok := keyvals[i].(Valuer); ok {
			return true
		}
	}
	return false
}

// Value is the return type of the Value method of the Valuer interface.
type Value interface{}

// A Valuer is able to convert itself into log Value.
//
// A Valuer passed to With stores a dynamic value which is reevaluated on each
// log event.
type Valuer interface {
	// Value returns a log Value instead of an interface{} to avoid
	// inadvertently matching types from packages not intended for use with
	// the gokit/log package.
	Value() Value
}

type timeStamp struct{}

func (*timeStamp) Value() Value {
	return time.Now()
}

// Timestamp is a Valuer that returns the result of time.Now() from its Value method.
var Timestamp = &timeStamp{}

type caller struct{}

func (*caller) Value() Value {
	return stack.Caller(3)
}

// Caller is a Valuer that returns a "gopkg.in/stack.v1".Call from its Value
// method.
var Caller = &caller{}
