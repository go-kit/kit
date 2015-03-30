package log

// Logger is the least-common-denominator interface for all log operations.
type Logger interface {
	With(keyvals ...interface{}) Logger
	Log(message string, keyvals ...interface{}) error
}
