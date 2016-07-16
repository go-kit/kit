package enabled

import "github.com/go-kit/kit/log/levels"

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
func New(logger levels.Levels, option Option) LevelEnabledLogger {
	l := LevelEnabledLogger{
		Levels: logger,
	}
	option(&l)
	return l
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

// Option sets a parameter for level enabled loggers.
type Option func(*LevelEnabledLogger)

// Debug enables all levels for the level enabled logger.
func Debug() Option {
	return func(l *LevelEnabledLogger) {
		l.debugEnabled = true
		l.infoEnabled = true
		l.warnEnabled = true
		l.errorEnabled = true
		l.critEnabled = true
	}
}

// Info enables info and above levels for the level enabled logger.
func Info() Option {
	return func(l *LevelEnabledLogger) {
		l.infoEnabled = true
		l.warnEnabled = true
		l.errorEnabled = true
		l.critEnabled = true
	}
}

// Warn enables warn and above levels for the level enabled logger.
func Warn() Option {
	return func(l *LevelEnabledLogger) {
		l.warnEnabled = true
		l.errorEnabled = true
		l.critEnabled = true
	}
}

// Error enables error and above levels for the level enabled logger.
func Error() Option {
	return func(l *LevelEnabledLogger) {
		l.errorEnabled = true
		l.critEnabled = true
	}
}

// Crit enables crit level for the level enabled logger.
func Crit() Option {
	return func(l *LevelEnabledLogger) {
		l.critEnabled = true
	}
}
