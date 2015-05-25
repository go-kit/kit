// Package log provides basic interfaces for structured logging.
//
// The fundamental interface is Logger. Loggers create log events from
// key/value data.
package log

import "sync/atomic"

// Logger is the fundamental interface for all log operations. Implementations
// must be safe for concurrent use by multiple goroutines. Log creates a log
// event from keyvals, a variadic sequence of alternating keys and values.
type Logger interface {
	Log(keyvals ...interface{}) error
}

// With returns a new Logger that includes keyvals in all log events. The
// returned Logger replaces all value elements (odd indexes) containing a
// Valuer with their generated value for each call to its Log method.
func With(logger Logger, keyvals ...interface{}) Logger {
	w, ok := logger.(*withLogger)
	if !ok {
		w = &withLogger{logger: logger}
	}
	return w.with(keyvals...)
}

type withLogger struct {
	logger    Logger
	keyvals   []interface{}
	hasValuer bool
}

func (l *withLogger) Log(keyvals ...interface{}) error {
	kvs := append(l.keyvals, keyvals...)
	if l.hasValuer {
		bindValues(kvs[:len(l.keyvals)])
	}
	return l.logger.Log(kvs...)
}

func (l *withLogger) with(keyvals ...interface{}) Logger {
	// Limiting the capacity of the stored keyvals ensures that a new
	// backing array is created if the slice must grow in Log or With.
	// Using the extra capacity without copying risks a data race that
	// would violate the Logger interface contract.
	n := len(l.keyvals) + len(keyvals)
	return &withLogger{
		logger:    l.logger,
		keyvals:   append(l.keyvals, keyvals...)[:n:n],
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
