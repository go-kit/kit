package level

import "github.com/go-kit/kit/log"

// Error returns a logger that includes an error level keyval.
func Error(logger log.Logger) log.Logger {
	return log.NewContext(logger).WithPrefix(levelKey, errorLevelValue)
}

// Warn returns a logger that includes a warn level keyval.
func Warn(logger log.Logger) log.Logger {
	return log.NewContext(logger).WithPrefix(levelKey, warnLevelValue)
}

// Info returns a logger that includes an info level keyval.
func Info(logger log.Logger) log.Logger {
	return log.NewContext(logger).WithPrefix(levelKey, infoLevelValue)
}

// Debug returns a logger that includes a debug level keyval.
func Debug(logger log.Logger) log.Logger {
	return log.NewContext(logger).WithPrefix(levelKey, debugLevelValue)
}

// NewFilter wraps next and implements level filtering. See the commentary on
// the Option functions for a detailed description of how to configure levels.
// If no options are provided, all leveled log events created with Debug,
// Info, Warn or Error helper methods are squelched and non-leveled log
// events are passed to next unmodified.
func NewFilter(next log.Logger, options ...Option) log.Logger {
	l := &logger{
		next: next,
	}
	for _, option := range options {
		option(l)
	}
	return l
}

type logger struct {
	next           log.Logger
	allowed        level
	squelchNoLevel bool
	errNotAllowed  error
	errNoLevel     error
}

func (l *logger) Log(keyvals ...interface{}) error {
	var hasLevel, levelAllowed bool
	for i := 1; i < len(keyvals); i += 2 {
		if v, ok := keyvals[i].(*levelValue); ok {
			hasLevel = true
			levelAllowed = l.allowed&v.level != 0
			break
		}
	}
	if !hasLevel && l.squelchNoLevel {
		return l.errNoLevel
	}
	if hasLevel && !levelAllowed {
		return l.errNotAllowed
	}
	return l.next.Log(keyvals...)
}

// Option sets a parameter for the leveled logger.
type Option func(*logger)

// AllowAll is an alias for AllowDebug.
func AllowAll() Option {
	return AllowDebug()
}

// AllowDebug allows error, warn, info and debug level log events to pass.
func AllowDebug() Option {
	return allowed(levelError | levelWarn | levelInfo | levelDebug)
}

// AllowInfo allows error, warn and info level log events to pass.
func AllowInfo() Option {
	return allowed(levelError | levelWarn | levelInfo)
}

// AllowWarn allows error and warn level log events to pass.
func AllowWarn() Option {
	return allowed(levelError | levelWarn)
}

// AllowError allows only error level log events to pass.
func AllowError() Option {
	return allowed(levelError)
}

// AllowNone allows no leveled log events to pass.
func AllowNone() Option {
	return allowed(0)
}

func allowed(allowed level) Option {
	return func(l *logger) { l.allowed = allowed }
}

// ErrNotAllowed sets the error to return from Log when it squelches a log
// event disallowed by the configured Allow[Level] option. By default,
// ErrNotAllowed is nil; in this case the log event is squelched with no
// error.
func ErrNotAllowed(err error) Option {
	return func(l *logger) { l.errNotAllowed = err }
}

// SquelchNoLevel instructs Log to squelch log events with no level, so that
// they don't proceed through to the wrapped logger. If SquelchNoLevel is set
// to true and a log event is squelched in this way, the error value
// configured with ErrNoLevel is returned to the caller.
func SquelchNoLevel(squelch bool) Option {
	return func(l *logger) { l.squelchNoLevel = squelch }
}

// ErrNoLevel sets the error to return from Log when it squelches a log event
// with no level. By default, ErrNoLevel is nil; in this case the log event is
// squelched with no error.
func ErrNoLevel(err error) Option {
	return func(l *logger) { l.errNoLevel = err }
}

var (
	levelKey        interface{} = "level"
	errorLevelValue             = &levelValue{level: levelError, name: "error"}
	warnLevelValue              = &levelValue{level: levelWarn, name: "warn"}
	infoLevelValue              = &levelValue{level: levelInfo, name: "info"}
	debugLevelValue             = &levelValue{level: levelDebug, name: "debug"}
)

type level byte

const (
	levelDebug level = 1 << iota
	levelInfo
	levelWarn
	levelError
)

type levelValue struct {
	name string
	level
}

func (v *levelValue) String() string {
	return v.name
}
