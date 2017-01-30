package level

import "github.com/go-kit/kit/log"

var (
	// Alternately, we could use a similarly inert logger that does nothing but
	// return a given error value.
	nop = log.NewNopLogger()

	// Invoking a leveling function with a Logger that neither
	// originated from nor wraps a Logger that originated from one of
	// the level-filtering factory functions still yields a
	// level-stamped Context, as if no filtering is in effect.
	defaultLeveler = &debugAndAbove{}
)

type leveler interface {
	Debug(log.Logger) log.Logger
	Info(log.Logger) log.Logger
	Warn(log.Logger) log.Logger
	Error(log.Logger) log.Logger
}

type leveledLogger struct {
	log.Logger
	leveler
}

func outermostLevelerOr(logger log.Logger, otherwise leveler) leveler {
	for {
		switch l := logger.(type) {
		case *leveledLogger:
			return l.leveler
			// Optimize unwrapping a Context by saving a type comparison.
		case *log.Context:
			logger = l.Delegate()
		default:
			logger = log.Delegate(logger)
		}
		if logger == nil {
			return otherwise
		}
	}
}

func outermostLeveler(logger log.Logger) leveler {
	return outermostLevelerOr(logger, nil)
}

func outermostEffectiveLeveler(logger log.Logger) leveler {
	return outermostLevelerOr(logger, defaultLeveler)
}

// Debug returns a logger ready to emit log records at the "debug"
// level, intended for fine-level detailed tracing information. If the
// supplied logger disallows records at that level, it instead returns
// an inert logger that drops the record.
func Debug(logger log.Logger) log.Logger {
	return outermostEffectiveLeveler(logger).Debug(logger)
}

// Info returns a logger ready to emit log records at the "info"
// level, intended for informational messages. If the supplied logger
// disallows records at that level, it instead returns an inert logger
// that drops the record.
func Info(logger log.Logger) log.Logger {
	return outermostEffectiveLeveler(logger).Info(logger)
}

// Warn returns a logger ready to emit log records at the "warn"
// level, intended for indicating potential problems. If the supplied
// logger disallows records at that level, it instead returns an inert
// logger that drops the record.
func Warn(logger log.Logger) log.Logger {
	return outermostEffectiveLeveler(logger).Warn(logger)
}

// Error returns a logger ready to emit log records at the "error"
// level, intended for indicating serious failures. If the supplied
// logger disallows records at that level, it instead returns an inert
// logger that drops the record.
func Error(logger log.Logger) log.Logger {
	return outermostEffectiveLeveler(logger).Error(logger)
}

func withLevel(level string, logger log.Logger) log.Logger {
	return log.NewContext(logger).With("level", level)
}

type debugAndAbove struct {
}

func (l debugAndAbove) Debug(logger log.Logger) log.Logger {
	return withLevel("debug", logger)
}

func (l debugAndAbove) Info(logger log.Logger) log.Logger {
	return withLevel("info", logger)
}

func (l debugAndAbove) Warn(logger log.Logger) log.Logger {
	return withLevel("warn", logger)
}

func (l debugAndAbove) Error(logger log.Logger) log.Logger {
	return withLevel("error", logger)
}

type infoAndAbove struct {
	debugAndAbove
}

func (infoAndAbove) Debug(logger log.Logger) log.Logger {
	return nop
}

type warnAndAbove struct {
	infoAndAbove
}

func (warnAndAbove) Info(logger log.Logger) log.Logger {
	return nop
}

type errorOnly struct {
	warnAndAbove
}

func (errorOnly) Warn(logger log.Logger) log.Logger {
	return nop
}

type none struct {
	errorOnly
}

func (none) Error(logger log.Logger) log.Logger {
	return nop
}

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
	if outermostLeveler(logger) != nil {
		return logger
	}
	return &leveledLogger{logger, debugAndAbove{}}
}

// AllowingInfoAndAbove returns a logger allowed to emit log records
// at levels "info" and above, dropping "debug"-level records, unless
// the supplied logger is already restricted to some narrower set of
// levels, in which case it retains that restriction.
func AllowingInfoAndAbove(logger log.Logger) log.Logger {
	switch outermostLeveler(logger).(type) {
	case infoAndAbove, warnAndAbove, errorOnly, none:
		return logger
	default:
		return &leveledLogger{logger, infoAndAbove{}}
	}
}

// AllowingWarnAndAbove returns a logger allowed to emit log records
// at levels "warn" and above, dropping "debug"- and "info"-level
// records, unless the supplied logger is already restricted to some
// narrower set of levels, in which case it retains that restriction.
func AllowingWarnAndAbove(logger log.Logger) log.Logger {
	switch outermostLeveler(logger).(type) {
	case warnAndAbove, errorOnly, none:
		return logger
	default:
		return &leveledLogger{logger, warnAndAbove{}}
	}
}

// AllowingErrorOnly returns a logger allowed to emit log records only
// at level "error", dropping "debug"-, "info"-, and "warn"-level
// records, unless the supplied logger is already restricted to some
// narrower set of levels, in which case it retains that restriction.
func AllowingErrorOnly(logger log.Logger) log.Logger {
	switch outermostLeveler(logger).(type) {
	case errorOnly, none:
		return logger
	default:
		return &leveledLogger{logger, errorOnly{}}
	}
}

// AllowingNone returns a logger that drops log records at all levels.
func AllowingNone(logger log.Logger) log.Logger {
	switch outermostLeveler(logger).(type) {
	case none:
		return logger
	default:
		return &leveledLogger{logger, none{}}
	}
}
