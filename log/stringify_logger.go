package log

import "fmt"

// StringifyLogger stringifies every value to make it printable by logfmt.
//
// Example:
//	Logger := log.LogfmtLogger(os.Stderr)
//	Logger = log.StringifyLogger{Logger}
type StringifyLogger struct {
	Logger
}

func (l StringifyLogger) Log(keyvals ...interface{}) error {
	for i := 1; i < len(keyvals); i += 2 {
		switch keyvals[i].(type) {
		case string, fmt.Stringer, fmt.Formatter:
		case error:
		default:
			keyvals[i] = StringWrap{Value: keyvals[i]}
		}
	}
	return l.Logger.Log(keyvals...)
}

var _ = fmt.Stringer(StringWrap{})

// StringWrap wraps the Value as a fmt.Stringer.
type StringWrap struct {
	Value interface{}
}

// String returns a string representation (%v) of the underlying Value.
func (sw StringWrap) String() string {
	return fmt.Sprintf("%v", sw.Value)
}
