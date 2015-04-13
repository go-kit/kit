package log

import (
	"io"
	"regexp"
)

// StdlibWriter wraps a Logger and allows it to be passed to the stdlib
// logger's SetOutput. It will extract date/timestamps, filenames, and
// messages, and place them under relevant keys.
type StdlibWriter struct {
	Logger
	timestampKey string
	fileKey      string
	messageKey   string
}

// StdlibWriterOption sets a parameter for the StdlibWriter.
type StdlibWriterOption func(*StdlibWriter)

// TimestampKey sets the key for the timestamp field. By default, it's "ts".
func TimestampKey(key string) StdlibWriterOption {
	return func(w *StdlibWriter) { w.timestampKey = key }
}

// FileKey sets the key for the file and line field. By default, it's "file".
func FileKey(key string) StdlibWriterOption {
	return func(w *StdlibWriter) { w.fileKey = key }
}

// MessageKey sets the key for the actual log message. By default, it's "msg".
func MessageKey(key string) StdlibWriterOption {
	return func(w *StdlibWriter) { w.messageKey = key }
}

// NewStdlibWriter returns a new StdlibWriter wrapper around the passed
// logger. It's designed to be passed to log.SetOutput.
func NewStdlibWriter(logger Logger, options ...StdlibWriterOption) io.Writer {
	w := StdlibWriter{
		Logger:       logger,
		timestampKey: "ts",
		fileKey:      "file",
		messageKey:   "msg",
	}
	for _, option := range options {
		option(&w)
	}
	return w
}

func (w StdlibWriter) Write(p []byte) (int, error) {
	result := subexps(p)
	keyvals := []interface{}{}
	var timestamp string
	if date, ok := result["date"]; ok && date != "" {
		timestamp = date
	}
	if time, ok := result["time"]; ok && time != "" {
		if timestamp != "" {
			timestamp += " "
		}
		timestamp += time
	}
	if timestamp != "" {
		keyvals = append(keyvals, w.timestampKey, timestamp)
	}
	if file, ok := result["file"]; ok && file != "" {
		keyvals = append(keyvals, w.fileKey, file)
	}
	if msg, ok := result["msg"]; ok {
		keyvals = append(keyvals, w.messageKey, msg)
	}
	if err := w.Logger.Log(keyvals...); err != nil {
		return 0, err
	}
	return len(p), nil
}

const (
	logRegexpDate = `(?P<date>[0-9]{4}/[0-9]{2}/[0-9]{2})?[ ]?`
	logRegexpTime = `(?P<time>[0-9]{2}:[0-9]{2}:[0-9]{2}(\.[0-9]+)?)?[ ]?`
	logRegexpFile = `(?P<file>[^:]+:[0-9]+)?`
	logRegexpMsg  = `(: )?(?P<msg>.*)`
)

var (
	logRegexp = regexp.MustCompile(logRegexpDate + logRegexpTime + logRegexpFile + logRegexpMsg)
)

func subexps(line []byte) map[string]string {
	m := logRegexp.FindSubmatch(line)
	if len(m) < len(logRegexp.SubexpNames()) {
		return map[string]string{}
	}
	result := map[string]string{}
	for i, name := range logRegexp.SubexpNames() {
		result[name] = string(m[i])
	}
	return result
}
