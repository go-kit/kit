package level

import (
	"github.com/go-kit/kit/log"
)

var (
	// LevelKey is the key part of a level keyval.
	LevelKey = "level"

	// ErrorLevelValue is the val part of an error-level keyval.
	ErrorLevelValue = "error"

	// WarnLevelValue is the val part of a warn-level keyval.
	WarnLevelValue = "warn"

	// InfoLevelValue is the val part of an info-level keyval.
	InfoLevelValue = "info"

	// DebugLevelValue is the val part of a debug-level keyval.
	DebugLevelValue = "debug"
)

var (
	// AllowAll is an alias for AllowDebugAndAbove.
	AllowAll = AllowDebugAndAbove

	// AllowDebugAndAbove allows all of the four default log levels.
	// It may be provided as the Allowed parameter in the Config struct.
	AllowDebugAndAbove = []string{ErrorLevelValue, WarnLevelValue, InfoLevelValue, DebugLevelValue}

	// AllowInfoAndAbove allows the default info, warn, and error log levels.
	// It may be provided as the Allowed parameter in the Config struct.
	AllowInfoAndAbove = []string{ErrorLevelValue, WarnLevelValue, InfoLevelValue}

	// AllowWarnAndAbove allows the default warn and error log levels.
	// It may be provided as the Allowed parameter in the Config struct.
	AllowWarnAndAbove = []string{ErrorLevelValue, WarnLevelValue}

	// AllowErrorOnly allows only the default error log level.
	// It may be provided as the Allowed parameter in the Config struct.
	AllowErrorOnly = []string{ErrorLevelValue}

	// AllowNone allows none of the default log levels.
	// It may be provided as the Allowed parameter in the Config struct.
	AllowNone = []string{}
)

// Error returns a logger with the LevelKey set to ErrorLevelValue.
func Error(logger log.Logger) log.Logger {
	return log.NewContext(logger).With(LevelKey, ErrorLevelValue)
}

// Warn returns a logger with the LevelKey set to WarnLevelValue.
func Warn(logger log.Logger) log.Logger {
	return log.NewContext(logger).With(LevelKey, WarnLevelValue)
}

// Info returns a logger with the LevelKey set to InfoLevelValue.
func Info(logger log.Logger) log.Logger {
	return log.NewContext(logger).With(LevelKey, InfoLevelValue)
}

// Debug returns a logger with the LevelKey set to DebugLevelValue.
func Debug(logger log.Logger) log.Logger {
	return log.NewContext(logger).With(LevelKey, DebugLevelValue)
}

// Config parameterizes the leveled logger.
type Config struct {
	// Allowed enumerates the accepted log levels. If a log event is encountered
	// with a LevelKey set to a value that isn't explicitly allowed, the event
	// will be squelched, and ErrSquelched returned.
	Allowed []string

	// ErrSquelched is returned to the caller when Log is invoked with a
	// LevelKey that hasn't been explicitly allowed. By default, ErrSquelched is
	// nil; in this case, the log event is squelched with no error.
	ErrSquelched error

	// AllowNoLevel will allow log events with no LevelKey to proceed through to
	// the wrapped logger without error. By default, log events with no LevelKey
	// will be squelched, and ErrNoLevel returned.
	AllowNoLevel bool

	// ErrNoLevel is returned to the caller when AllowNoLevel is false, and Log
	// is invoked without a LevelKey. By default, ErrNoLevel is nil; in this
	// case, the log event is squelched with no error.
	ErrNoLevel error
}

// New wraps the logger and implements level checking. See the commentary on the
// Config object for a detailed description of how to configure levels.
func New(next log.Logger, config Config) log.Logger {
	return &logger{
		next:         next,
		allowed:      makeSet(config.Allowed),
		errSquelched: config.ErrSquelched,
		allowNoLevel: config.AllowNoLevel,
		errNoLevel:   config.ErrNoLevel,
	}
}

type logger struct {
	next         log.Logger
	allowed      map[string]struct{}
	errSquelched error
	allowNoLevel bool
	errNoLevel   error
}

func (l *logger) Log(keyvals ...interface{}) error {
	var hasLevel, levelAllowed bool
	for i := 0; i < len(keyvals); i += 2 {
		if k, ok := keyvals[i].(string); !ok || k != LevelKey {
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
	if !hasLevel && !l.allowNoLevel {
		return l.errNoLevel
	}
	if hasLevel && !levelAllowed {
		return l.errSquelched
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
