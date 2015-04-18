// Package log provides basic interfaces for structured logging.
//
// The fundamental interface is Logger. Loggers create log events from
// key/value data.
package log

// Logger is the fundamental interface for all log operations.
//
// Log creates a log event from keyvals, a variadic sequence of alternating
// keys and values.
//
// Logger implementations must be safe for concurrent use by multiple
// goroutines.
type Logger interface {
	Log(keyvals ...interface{}) error
}

// With returns a new Logger that includes keyvals in all log events.
//
// If logger implements the Wither interface, the result of
// logger.With(keyvals...) is returned.
//
// When wrapping a basic Logger, With returns a logger that calls BindValues
// on the stored keyvals when logging.
func With(logger Logger, keyvals ...interface{}) Logger {
	w, ok := logger.(Wither)
	if !ok {
		w = &withLogger{logger: logger}
	}
	return w.With(keyvals...)
}

type withLogger struct {
	logger  Logger
	keyvals []interface{}
}

func (l *withLogger) Log(keyvals ...interface{}) error {
	return l.logger.Log(append(BindValues(l.keyvals...), keyvals...)...)
}

func (l *withLogger) With(keyvals ...interface{}) Logger {
	// Limiting the capacity of the stored keyvals ensures that a new
	// backing array is created if the slice must grow in Log or With.
	// Using the extra capacity without copying risks a data race that
	// would violate the Logger interface contract.
	n := len(l.keyvals) + len(keyvals)
	return &withLogger{
		logger:  l.logger,
		keyvals: append(l.keyvals, keyvals...)[:n:n],
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

// A Wither creates Loggers that include keyvals in all log events.
//
// The With function uses Wither if available.
//
// It is recommended that implementations of With call BindValues on stored
// keyvals before each log event.
type Wither interface {
	With(keyvals ...interface{}) Logger
}

// NewDiscardLogger returns a logger that does not log anything.
func NewDiscardLogger() Logger {
	return LoggerFunc(func(...interface{}) error {
		return nil
	})
}
