package log

// Logger is the least-common-denominator interface for all log operations.
type Logger interface {
	Log(keyvals ...interface{}) error
}

// With new, contextualized Logger with the passed keyvals already applied.
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

// Wither describes an optimization that Logger implementations may make. If a
// Logger implements Wither, the package-level With function will invoke it
// when creating a new, contextual logger.
type Wither interface {
	With(keyvals ...interface{}) Logger
}
