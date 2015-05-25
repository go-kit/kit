package log

// DefaultLogger is used by gokit components. By default, it's a LogfmtLogger
// that writes to the stdlib log.
var DefaultLogger = NewLogfmtLogger(StdlibWriter{})
