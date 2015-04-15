// Package log provides basic interfaces for structured logging.
//
// The fundamental interface is Logger. Loggers create log events from
// key/value data.
package log

// Logger is the fundamental interface for all log operations.
//
// Log creates a log event from keyvals, a variadic sequence of alternating
// keys and values.
type Logger interface {
	Log(keyvals ...interface{}) error
}

// With returns a new Logger that incluces keyvals in all log events.
//
// If logger implements the Wither interface, the result of
// logger.With(keyvals...) is returned.
func With(logger Logger, keyvals ...interface{}) Logger {
	if w, ok := logger.(Wither); ok {
		return w.With(keyvals...)
	}
	return LoggerFunc(func(kvs ...interface{}) error {
		return logger.Log(append(keyvals, kvs...)...)
	})
}

// LoggerFunc is an adapter to allow use of ordinary functions as Loggers. If
// f is a function with the appropriate signature, LoggerFunc(f) is a Logger
// object that calls f.
type LoggerFunc func(...interface{}) error

// Log implements Logger by calling f(keyvals...).
func (f LoggerFunc) Log(keyvals ...interface{}) error {
	return f(keyvals...)
}

// A Wither creates Loggers that include keyvals in all log events.
//
// The With function uses Wither if available.
type Wither interface {
	With(keyvals ...interface{}) Logger
}
