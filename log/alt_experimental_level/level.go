package level

import "github.com/go-kit/kit/log"

type leveler interface {
	Debug(log.Logger) log.Logger
	Info(log.Logger) log.Logger
	Warn(log.Logger) log.Logger
	Error(log.Logger) log.Logger
}

type level byte

const (
	levelDebug level = 1 << iota
	levelInfo
	levelWarn
	levelError
)

type levelValue struct {
	level
	name string
}

func (v *levelValue) String() string {
	return v.name
}

var (
	valDebug = &levelValue{levelDebug, "debug"}
	valInfo  = &levelValue{levelInfo, "info"}
	valWarn  = &levelValue{levelWarn, "warn"}
	valError = &levelValue{levelError, "error"}
)

func withLevel(v *levelValue, logger log.Logger) *log.Context {
	return log.NewContext(logger).WithPrefix("level", v)
}

// Debug returns a logger ready to emit log records at the "debug"
// level, intended for fine-level detailed tracing information. If the
// supplied logger disallows records at that level, it instead returns
// an inert logger that drops the record.
func Debug(logger log.Logger) log.Logger {
	return withLevel(valDebug, logger)
}

// Info returns a logger ready to emit log records at the "info"
// level, intended for informational messages. If the supplied logger
// disallows records at that level, it instead returns an inert logger
// that drops the record.
func Info(logger log.Logger) log.Logger {
	return withLevel(valInfo, logger)
}

// Warn returns a logger ready to emit log records at the "warn"
// level, intended for indicating potential problems. If the supplied
// logger disallows records at that level, it instead returns an inert
// logger that drops the record.
func Warn(logger log.Logger) log.Logger {
	return withLevel(valWarn, logger)
}

// Error returns a logger ready to emit log records at the "error"
// level, intended for indicating serious failures. If the supplied
// logger disallows records at that level, it instead returns an inert
// logger that drops the record.
func Error(logger log.Logger) log.Logger {
	return withLevel(valError, logger)
}

func rejectLevelsOtherThan(mask level) log.Projection {
	return func(keyvals []interface{}) ([]interface{}, bool) {
		for i, end := 1, len(keyvals); i < end; i += 2 {
			if l, ok := keyvals[i].(*levelValue); ok {
				if l.level&mask == 0 {
					return nil, false
				}
				break
			}
		}
		return keyvals, true
	}
}

var (
	preserveInfoAndAbove = rejectLevelsOtherThan(levelInfo | levelWarn | levelError)
	preserveWarnAndAbove = rejectLevelsOtherThan(levelWarn | levelError)
	preserveErrorOnly    = rejectLevelsOtherThan(levelError)
)

// AllowingAll returns a logger allowed to emit log records at all
// levels, unless the supplied logger is already restricted to some
// narrower set of levels, in which case it retains that restriction.
//
// The behavior is equivalent to AllowingDebugAndAbove.
func AllowingAll(logger log.Logger) log.Logger {
	return AllowingDebugAndAbove(logger)
}

// AllowingDebugAndAbove returns a logger allowed to emit log records
// at all levels, unless the supplied logger is already restricted to
// some narrower set of levels, in which case it retains that
// restriction.
func AllowingDebugAndAbove(logger log.Logger) log.Logger {
	return logger
}

// AllowingInfoAndAbove returns a logger allowed to emit log records
// at levels "info" and above, dropping "debug"-level records, unless
// the supplied logger is already restricted to some narrower set of
// levels, in which case it retains that restriction.
func AllowingInfoAndAbove(logger log.Logger) log.Logger {
	return log.NewContext(logger).WithProjection(preserveInfoAndAbove)
}

// AllowingWarnAndAbove returns a logger allowed to emit log records
// at levels "warn" and above, dropping "debug"- and "info"-level
// records, unless the supplied logger is already restricted to some
// narrower set of levels, in which case it retains that restriction.
func AllowingWarnAndAbove(logger log.Logger) log.Logger {
	return log.NewContext(logger).WithProjection(preserveWarnAndAbove)
}

// AllowingErrorOnly returns a logger allowed to emit log records only
// at level "error", dropping "debug"-, "info"-, and "warn"-level
// records, unless the supplied logger is already restricted to some
// narrower set of levels, in which case it retains that restriction.
func AllowingErrorOnly(logger log.Logger) log.Logger {
	return log.NewContext(logger).WithProjection(preserveErrorOnly)
}

// AllowingNone returns a logger that drops log records at all levels.
func AllowingNone(logger log.Logger) log.Logger {
	return log.NewContext(logger).WithProjection(func(keyvals []interface{}) ([]interface{}, bool) {
		for i, end := 1, len(keyvals); i < end; i += 2 {
			if _, ok := keyvals[i].(*levelValue); ok {
				return nil, false
			}
		}
		return keyvals, true
	})
}
