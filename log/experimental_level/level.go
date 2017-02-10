package level

import (
	"github.com/go-kit/kit/log"
)

var (
	levelKey        = "level"
	errorLevelValue = "error"
	warnLevelValue  = "warn"
	infoLevelValue  = "info"
	debugLevelValue = "debug"
)

// AllowAll is an alias for AllowDebugAndAbove.
func AllowAll() []string {
	return AllowDebugAndAbove()
}

// AllowDebugAndAbove allows all of the four default log levels.
// Its return value may be provided with the Allowed Option.
func AllowDebugAndAbove() []string {
	return []string{errorLevelValue, warnLevelValue, infoLevelValue, debugLevelValue}
}

// AllowInfoAndAbove allows the default info, warn, and error log levels.
// Its return value may be provided with the Allowed Option.
func AllowInfoAndAbove() []string {
	return []string{errorLevelValue, warnLevelValue, infoLevelValue}
}

// AllowWarnAndAbove allows the default warn and error log levels.
// Its return value may be provided with the Allowed Option.
func AllowWarnAndAbove() []string {
	return []string{errorLevelValue, warnLevelValue}
}

// AllowErrorOnly allows only the default error log level.
// Its return value may be provided with the Allowed Option.
func AllowErrorOnly() []string {
	return []string{errorLevelValue}
}

// AllowNone allows none of the default log levels.
// Its return value may be provided with the Allowed Option.
func AllowNone() []string {
	return []string{}
}

// Error returns a logger with the level key set to ErrorLevelValue.
func Error(logger log.Logger) log.Logger {
	return log.NewContext(logger).WithPrefix(levelKey, errorLevelValue)
}

// Warn returns a logger with the level key set to WarnLevelValue.
func Warn(logger log.Logger) log.Logger {
	return log.NewContext(logger).WithPrefix(levelKey, warnLevelValue)
}

// Info returns a logger with the level key set to InfoLevelValue.
func Info(logger log.Logger) log.Logger {
	return log.NewContext(logger).WithPrefix(levelKey, infoLevelValue)
}

// Debug returns a logger with the level key set to DebugLevelValue.
func Debug(logger log.Logger) log.Logger {
	return log.NewContext(logger).WithPrefix(levelKey, debugLevelValue)
}

// New wraps the logger and implements level checking. See the commentary on the
// Option functions for a detailed description of how to configure levels.
// If no options are provided, all leveled log events created with level.Debug,
// Info, Warn or Error helper methods will be squelched.
func New(next log.Logger, options ...Option) log.Logger {
	l := logger{
		next: next,
	}
	for _, option := range options {
		option(&l)
	}
	return &l
}

// Allowed enumerates the accepted log levels. If a log event is encountered
// with a level key set to a value that isn't explicitly allowed, the event
// will be squelched, and ErrNotAllowed returned.
func Allowed(allowed []string) Option {
	return func(l *logger) { l.allowed = makeSet(allowed) }
}

// ErrNoLevel is returned to the caller when SquelchNoLevel is true, and Log
// is invoked without a level key. By default, ErrNoLevel is nil; in this
// case, the log event is squelched with no error.
func ErrNotAllowed(err error) Option {
	return func(l *logger) { l.errNotAllowed = err }
}

// SquelchNoLevel will squelch log events with no level key, so that they
// don't proceed through to the wrapped logger. If SquelchNoLevel is set to
// true and a log event is squelched in this way, ErrNoLevel is returned to
// the caller.
func SquelchNoLevel(squelch bool) Option {
	return func(l *logger) { l.squelchNoLevel = squelch }
}

// ErrNoLevel is returned to the caller when SquelchNoLevel is true, and Log
// is invoked without a level key. By default, ErrNoLevel is nil; in this
// case, the log event is squelched with no error.
func ErrNoLevel(err error) Option {
	return func(l *logger) { l.errNoLevel = err }
}

// Option sets a parameter for the leveled logger.
type Option func(*logger)

type logger struct {
	next           log.Logger
	allowed        map[string]struct{}
	errNotAllowed  error
	squelchNoLevel bool
	errNoLevel     error
}

func (l *logger) Log(keyvals ...interface{}) error {
	var hasLevel, levelAllowed bool
	for i := 0; i < len(keyvals); i += 2 {
		if k, ok := keyvals[i].(string); !ok || k != levelKey {
			continue
		}
		hasLevel = true
		if i >= len(keyvals) {
			continue
		}
		v, ok := keyvals[i+1].(string)
		if !ok {
			continue
		}
		_, levelAllowed = l.allowed[v]
		break
	}
	if !hasLevel && l.squelchNoLevel {
		return l.errNoLevel
	}
	if hasLevel && !levelAllowed {
		return l.errNotAllowed
	}
	return l.next.Log(keyvals...)
}

func makeSet(a []string) map[string]struct{} {
	m := make(map[string]struct{}, len(a))
	for _, s := range a {
		m[s] = struct{}{}
	}
	return m
}
