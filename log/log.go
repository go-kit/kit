package log

// Logger is the least-common-denominator interface for all log operations.
type Logger interface {
	With(Field) Logger
	Logf(string, ...interface{}) error
}

// Field is a key/value pair associated with a log event. Fields may be
// ignored by implementations.
type Field struct {
	Key   string
	Value interface{}
}
