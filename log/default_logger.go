package log

// DefaultLogger is used by gokit components. By default, it's a PrefixLogger
// that writes to the stdlib log.
var DefaultLogger = NewPrefixLogger(StdlibWriter{})
