package enabled

import (
	"errors"

	"github.com/go-kit/kit/log/levels"
)

// LevelEnabledLogger provides a wrapper around a logger with a feature
// for checking a level is enabled.  It has five
// levels: debug, info, warning (warn), error, and critical (crit).
//
// It is users' responsibility not to call the Log method of a logger.
type LevelEnabledLogger struct {
	levels.Levels
	debugEnabled bool
	infoEnabled  bool
	warnEnabled  bool
	errorEnabled bool
	critEnabled  bool
}

// New creates a new level enabled logger, wrapping the passed logger.
// level must be one of debug, info, warning (warn), error, critical (crit)
// or an empty string (which means info).
// It returns an error only when the level value is invalid.
func New(logger levels.Levels, level string) (LevelEnabledLogger, error) {
	switch level {
	case "debug":
		return NewDebug(logger), nil
	case "", "info":
		return NewInfo(logger), nil
	case "warn", "warning":
		return NewWarn(logger), nil
	case "error":
		return NewError(logger), nil
	case "crit", "critical":
		return NewCrit(logger), nil
	default:
		return LevelEnabledLogger{}, errors.New("invalid log level")
	}
}

// NewDebug creates a new debug level enabled logger, wrapping the passed logger.
func NewDebug(logger levels.Levels) LevelEnabledLogger {
	return LevelEnabledLogger{
		Levels:       logger,
		debugEnabled: true,
		infoEnabled:  true,
		warnEnabled:  true,
		errorEnabled: true,
		critEnabled:  true,
	}
}

// NewInfo creates a new info level enabled logger, wrapping the passed logger.
func NewInfo(logger levels.Levels) LevelEnabledLogger {
	return LevelEnabledLogger{
		Levels:       logger,
		infoEnabled:  true,
		warnEnabled:  true,
		errorEnabled: true,
		critEnabled:  true,
	}
}

// NewWarn creates a new warn level enabled logger, wrapping the passed logger.
func NewWarn(logger levels.Levels) LevelEnabledLogger {
	return LevelEnabledLogger{
		Levels:       logger,
		warnEnabled:  true,
		errorEnabled: true,
		critEnabled:  true,
	}
}

// NewError creates a new error level enabled logger, wrapping the passed logger.
func NewError(logger levels.Levels) LevelEnabledLogger {
	return LevelEnabledLogger{
		Levels:       logger,
		errorEnabled: true,
		critEnabled:  true,
	}
}

// NewCrit creates a new crit level enabled logger, wrapping the passed logger.
func NewCrit(logger levels.Levels) LevelEnabledLogger {
	return LevelEnabledLogger{
		Levels:       logger,
		errorEnabled: true,
		critEnabled:  true,
	}
}

// DebugEnabled returns the debug level is enabled
func (l LevelEnabledLogger) DebugEnabled() bool { return l.debugEnabled }

// InfoEnabled returns the info level is enabled
func (l LevelEnabledLogger) InfoEnabled() bool { return l.infoEnabled }

// WarnEnabled returns the warn level is enabled
func (l LevelEnabledLogger) WarnEnabled() bool { return l.warnEnabled }

// ErrorEnabled returns the error level is enabled
func (l LevelEnabledLogger) ErrorEnabled() bool { return l.errorEnabled }

// CritEnabled returns the crit level is enabled
func (l LevelEnabledLogger) CritEnabled() bool { return l.critEnabled }
