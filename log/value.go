package log

import (
	"time"

	"gopkg.in/stack.v1"
)

// A Valuer generates a log value. When passed to With, it represents a
// dynamic value which is re-evaluated with each log event.
type Valuer func() interface{}

// BindValues replaces all value elements (odd indexes) containing a Valuer
// with their generated value.
func BindValues(keyvals []interface{}) {
	for i := 1; i < len(keyvals); i += 2 {
		if v, ok := keyvals[i].(Valuer); ok {
			keyvals[i] = v()
		}
	}
}

// ContainsValuer returns true if any of the value elements (odd indexes)
// contain a Valuer.
func ContainsValuer(keyvals []interface{}) bool {
	for i := 1; i < len(keyvals); i += 2 {
		if _, ok := keyvals[i].(Valuer); ok {
			return true
		}
	}
	return false
}

// Timestamp returns a Valuer that invokes the underlying function when bound,
// returning a time.Time. Users will probably want to use DefaultTimestamp or
// DefaultTimestampUTC.
func Timestamp(t func() time.Time) Valuer {
	return func() interface{} { return t() }
}

var (
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

// Caller returns a Valuer that returns a file and line from a specified depth
// in the callstack. Users will probably want to use DefaultCaller.
func Caller(depth int) Valuer {
	return func() interface{} { return stack.Caller(depth) }
}

var (
	// DefaultCaller is a Valuer that returns the file and line where the Log
	// method was invoked.
	DefaultCaller = Caller(3)
)
