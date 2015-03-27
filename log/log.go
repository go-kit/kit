package log

// Logger is the least-common-denominator interface for all log operations.
type Logger interface {
	With(...Field) Logger
	Log(string) error
}

// Field is a key/value pair associated with a log event.
type Field struct {
	Key   string
	Value interface{}
}
