package log

import (
	"time"

	"gopkg.in/stack.v1"
)

// A Valuer generates a log value. When passed to With, it represents a
// dynamic value which is re-evaluated with each log event.
type Valuer func() interface{}

// BindValues returns a slice with all value elements (odd indexes) containing
// a Valuer replaced by their generated value. If no Valuers are found, the
// original slice is returned.
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
	// Timestamp returns a Valuer that invokes the underlying function when
	// bound, returning a time.Time. Users will probably want to use
	// DefaultTimestamp or DefaultTimestampUTC.
	Timestamp = func(t func() time.Time) Valuer {
		return func() interface{} { return t() }
	}

	// DefaultTimestamp is a Timestamp Valuer that returns the current
	// wallclock time, respecting time zones, when bound.
	DefaultTimestamp = Timestamp(time.Now)

	// DefaultTimestampUTC wraps DefaultTimestamp but ensures the returned
	// time is always in UTC. Note that it invokes DefaultTimestamp, and so
	// reflects any changes to the DefaultTimestamp package global.
	DefaultTimestampUTC = Timestamp(func() time.Time {
		return DefaultTimestamp().(time.Time).UTC()
	})
)

var (
	// Caller is a Valuer that returns a file and line from a specified depth
	// in the callstack. Users will probably want to use DefaultCaller.
	Caller = func(depth int) Valuer {
		return func() interface{} { return stack.Caller(depth) }
	}

	// DefaultCaller is a Valuer that returns the file and line where the Log
	// method was invoked.
	DefaultCaller = Caller(3)
)
