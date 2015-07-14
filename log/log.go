// Package log provides basic interfaces for structured logging.
//
// The fundamental interface is Logger. Loggers create log events from
// key/value data.
package log

import "sync/atomic"

// Logger is the fundamental interface for all log operations. Log creates a
// log event from keyvals, a variadic sequence of alternating keys and values.
// Implementations must be safe for concurrent use by multiple goroutines. In
// particular, any implementation of Logger that appends to keyvals or
// modifies any of its elements must make a copy first.
type Logger interface {
	Log(keyvals ...interface{}) error
}

// NewContext returns a new Context that logs to logger.
func NewContext(logger Logger) Context {
	if c, ok := logger.(Context); ok {
		return c
	}
	return Context{logger: logger}
}

// A Context wraps a Logger and holds keyvals that it includes in all log
// events. When logging, a Context replaces all value elements (odd indexes)
// containing a Valuer with their generated value for each call to its Log
// method.
type Context struct {
	logger    Logger
	keyvals   []interface{}
	hasValuer bool
}

// Log replaces all value elements (odd indexes) containing a Valuer in the
// stored context with their generated value, appends keyvals, and passes the
// result to the wrapped Logger.
func (l Context) Log(keyvals ...interface{}) error {
	if len(keyvals)%2 != 0 {
		panic("bad keyvals")
	}
	kvs := append(l.keyvals, keyvals...)
	if l.hasValuer {
		// If no keyvals were appended above then we must copy l.keyvals so
		// that future log events will reevaluate the stored Valuers.
		if len(keyvals) == 0 {
			kvs = append([]interface{}{}, l.keyvals...)
		}
		bindValues(kvs[:len(l.keyvals)])
	}
	return l.logger.Log(kvs...)
}

// With returns a new Context with keyvals appended to those of the receiver.
func (l Context) With(keyvals ...interface{}) Context {
	if len(keyvals) == 0 {
		return l
	}
	if len(keyvals)%2 != 0 {
		panic("bad keyvals")
	}
	// Limiting the capacity of the stored keyvals ensures that a new
	// backing array is created if the slice must grow in Log or With.
	// Using the extra capacity without copying risks a data race that
	// would violate the Logger interface contract.
	n := len(l.keyvals) + len(keyvals)
	return Context{
		logger:    l.logger,
		keyvals:   append(l.keyvals, keyvals...)[:n:n],
		hasValuer: l.hasValuer || containsValuer(keyvals),
	}
}

// WithPrefix returns a new Context with keyvals prepended to those of the
// receiver.
func (l Context) WithPrefix(keyvals ...interface{}) Context {
	if len(keyvals) == 0 {
		return l
	}
	if len(keyvals)%2 != 0 {
		panic("bad keyvals")
	}
	// Limiting the capacity of the stored keyvals ensures that a new
	// backing array is created if the slice must grow in Log or With.
	// Using the extra capacity without copying risks a data race that
	// would violate the Logger interface contract.
	n := len(l.keyvals) + len(keyvals)
	kvs := make([]interface{}, 0, n)
	kvs = append(kvs, keyvals...)
	kvs = append(kvs, l.keyvals...)
	return Context{
		logger:    l.logger,
		keyvals:   kvs,
		hasValuer: l.hasValuer || containsValuer(keyvals),
	}
}

// LoggerFunc is an adapter to allow use of ordinary functions as Loggers. If
// f is a function with the appropriate signature, LoggerFunc(f) is a Logger
// object that calls f.
type LoggerFunc func(...interface{}) error

// Log implements Logger by calling f(keyvals...).
func (f LoggerFunc) Log(keyvals ...interface{}) error {
	return f(keyvals...)
}

// SwapLogger wraps another logger that may be safely replaced while other
// goroutines use the SwapLogger concurrently. The zero value for a SwapLogger
// will discard all log events without error.
//
// SwapLogger serves well as a package global logger that can be changed by
// importers.
type SwapLogger struct {
	logger atomic.Value
}

type loggerStruct struct {
	Logger
}

// Log implements the Logger interface by forwarding keyvals to the currently
// wrapped logger. It does not log anything if the wrapped logger is nil.
func (l *SwapLogger) Log(keyvals ...interface{}) error {
	s, ok := l.logger.Load().(loggerStruct)
	if !ok || s.Logger == nil {
		return nil
	}
	return s.Log(keyvals...)
}

// Swap replaces the currently wrapped logger with logger. Swap may be called
// concurrently with calls to Log from other goroutines.
func (l *SwapLogger) Swap(logger Logger) {
	l.logger.Store(loggerStruct{logger})
}
