package log

import (
	"time"

	"gopkg.in/stack.v1"
)

// Value is the return type of the Value method of the Valuer interface.
type Value interface{}

// A Valuer generates a log value. When passed to With, it represents a
// dynamic value which is re-evaluated with each log event.
// It returns a log Value instead of an interface{} to avoid
// inadvertently matching types from packages not intended for use with
// gokit/log.
type Valuer func() Value

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
			bound[i] = v()
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

var (
	// Timestamp returns a Valuer that invokes the underlying function when bound,
	// returning a time.Time. Users will probably want to use DefaultTimestamp or
	// DefaultTimestampUTC.
	Timestamp = func(t func() time.Time) Valuer {
		return func() Value { return t() }
	}
	// DefaultTimestamp is a Timestamp Valuer that returns the current wallclock
	// time, respecting time zones, when bound.
	DefaultTimestamp = Timestamp(time.Now)
	// DefaultTimestampUTC wraps DefaultTimestamp but ensures the returned
	// time is always in UTC. Note that it invokes DefaultTimestamp, and so
	// reflects any changes to the DefaultTimestamp package global.
	DefaultTimestampUTC = Timestamp(func() time.Time {
		return DefaultTimestamp().(time.Time).UTC()
	})
	// Caller is a Valuer that returns a file and line from a specified depth in
	// the callstack. Users will probably want to use DefaultCaller.
	Caller = func(depth int) Valuer {
		return func() Value { return stack.Caller(depth) }
	}
	// DefaultCaller is a Valuer that returns the file and line where the Log
	// method was invoked.
	DefaultCaller = Caller(3)
)
