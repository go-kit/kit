package levels

import "github.com/go-kit/kit/log"

// Levels provides a leveled logging wrapper around a logger. It has five
// levels: debug, info, warning (warn), error, and critical (crit). If you
// want a different set of levels, you can create your own levels type very
// easily, and you can elide the configuration.
type Levels struct {
	ctx      *log.Context
	levelKey string

	// We have a choice between storing level values in string fields or
	// making a separate context for each level. When using string fields the
	// Log method must combine the base context, the level data, and the
	// logged keyvals; but the With method only requires updating one context.
	// If we instead keep a separate context for each level the Log method
	// must only append the new keyvals; but the With method would have to
	// update all five contexts.

	// Roughly speaking, storing multiple contexts breaks even if the ratio of
	// Log/With calls is more than the number of levels. We have chosen to
	// make the With method cheap and the Log method a bit more costly because
	// we do not expect most applications to Log more than five times for each
	// call to With.

	debugValue string
	infoValue  string
	warnValue  string
	errorValue string
	critValue  string

	// FilterLogLevel is an option allowing logs below a certain level to be
	// discarded rather than logged. For example, a consumer may set LogLevel
	// to levels.Warn, and any logs to Debug or Info would become no-ops.
	filterLogLevel LogLevel

	// Promote errors
	promoteErrors       bool
	errorKey            string
	promoteErrorToLevel LogLevel
}

// levelCommittedLogger embeds a log level that the user has selected and prepends
// the associated level string as a prefix to every log line. Log level may be
// affected by error promotion; and log lines may be suppressed based on the value
// of FilterLogLevel.
type levelCommittedLogger struct {
	levels Levels

	committedLevel LogLevel
}

type LogLevel int

var (
	// Debug should be used for information that is useful to developers for
	// forensic purposes but is not useful information for everyday operations.
	Debug LogLevel = 0

	// Info should be used for information that is useful and actionable in
	// production operations when everything is running smoothly.
	Info LogLevel = 1

	// Warn should be used to note that the system is still performing its job
	// successfully, but a notable expectation for proper operation was not
	// met. If left untreated, warnings could eventually escalate into errors.
	Warn LogLevel = 2

	// Error should be used to flag when the system has failed to uphold its
	// operational contract in some way, and the failure was not recoverable.
	Error LogLevel = 3

	// Crit should only be used when an error occurs that is so catastrophic,
	// the system is going to immediately become unavailable for any future
	// operations until the problem is repaired.
	Crit LogLevel = 4
)

// New creates a new leveled logger, wrapping the passed logger.
func New(logger log.Logger, options ...Option) Levels {
	l := Levels{
		ctx:      log.NewContext(logger),
		levelKey: "level",

		debugValue: "debug",
		infoValue:  "info",
		warnValue:  "warn",
		errorValue: "error",
		critValue:  "crit",

		filterLogLevel: Debug,

		promoteErrors:       false,
		errorKey:            "error",
		promoteErrorToLevel: Error,
	}
	for _, option := range options {
		option(&l)
	}
	return l
}

// With returns a new leveled logger that includes keyvals in all log events.
func (l Levels) With(keyvals ...interface{}) Levels {
	return Levels{
		ctx:        l.ctx.With(keyvals...),
		levelKey:   l.levelKey,
		debugValue: l.debugValue,
		infoValue:  l.infoValue,
		warnValue:  l.warnValue,
		errorValue: l.errorValue,
		critValue:  l.critValue,

		filterLogLevel: l.filterLogLevel,

		promoteErrors:       l.promoteErrors,
		errorKey:            l.errorKey,
		promoteErrorToLevel: l.promoteErrorToLevel,
	}
}

func (l levelCommittedLogger) With(keyvals ...interface{}) levelCommittedLogger {
	return levelCommittedLogger{
		levels:         l.levels.With(keyvals...),
		committedLevel: l.committedLevel,
	}
}

func (l levelCommittedLogger) Log(keyvals ...interface{}) error {
	lvl, ctx := l.committedLevel, l.levels.ctx.With(keyvals...)

	// Check whether the log level should be promoted because of an error.
	if l.levels.promoteErrors && lvl < l.levels.promoteErrorToLevel && ctx.HasValue(l.levels.errorKey) {
		lvl = l.levels.promoteErrorToLevel
	}

	// Suppress logging if the level of this log line is below the minimum
	// log level we want to see.
	if lvl < l.levels.filterLogLevel {
		return nil
	}

	// Get the string associated with the current logLevel.
	var levelValue string
	switch lvl {
	case Debug:
		levelValue = l.levels.debugValue
	case Info:
		levelValue = l.levels.infoValue
	case Warn:
		levelValue = l.levels.warnValue
	case Error:
		levelValue = l.levels.errorValue
	case Crit:
		levelValue = l.levels.critValue
	}

	return ctx.WithPrefix(l.levels.levelKey, levelValue).Log()
}

// Debug returns a debug level logger.
func (l Levels) Debug() log.Logger {
	return levelCommittedLogger{l, Debug}
}

// Info returns an info level logger.
func (l Levels) Info() log.Logger {
	return levelCommittedLogger{l, Info}
}

// Warn returns a warning level logger.
func (l Levels) Warn() log.Logger {
	return levelCommittedLogger{l, Warn}
}

// Error returns an error level logger.
func (l Levels) Error() log.Logger {
	return levelCommittedLogger{l, Error}
}

// Crit returns a critical level logger.
func (l Levels) Crit() log.Logger {
	return levelCommittedLogger{l, Crit}
}

// Option sets a parameter for leveled loggers.
type Option func(*Levels)

// Key sets the key for the field used to indicate log level. By default,
// the key is "level".
func Key(key string) Option {
	return func(l *Levels) { l.levelKey = key }
}

// DebugValue sets the value for the field used to indicate the debug log
// level. By default, the value is "debug".
func DebugValue(value string) Option {
	return func(l *Levels) { l.debugValue = value }
}

// InfoValue sets the value for the field used to indicate the info log level.
// By default, the value is "info".
func InfoValue(value string) Option {
	return func(l *Levels) { l.infoValue = value }
}

// WarnValue sets the value for the field used to indicate the warning log
// level. By default, the value is "warn".
func WarnValue(value string) Option {
	return func(l *Levels) { l.warnValue = value }
}

// ErrorValue sets the value for the field used to indicate the error log
// level. By default, the value is "error".
func ErrorValue(value string) Option {
	return func(l *Levels) { l.errorValue = value }
}

// CritValue sets the value for the field used to indicate the critical log
// level. By default, the value is "crit".
func CritValue(value string) Option {
	return func(l *Levels) { l.critValue = value }
}

// FilterLogLevel sets the value for the minimum log level that will be
// printed by the logger. By default, the value is levels.Debug.
func FilterLogLevel(value LogLevel) Option {
	return func(l *Levels) { l.filterLogLevel = value }
}

// PromoteErrors sets whether log lines with errors will be promoted to a
// higher log level. By default, the value is false.
func PromoteErrors(value bool) Option {
	return func(l *Levels) { l.promoteErrors = value }
}

// ErrorKey sets the key where errors will be stored in the log line. This
// is used if PromoteErrors is set to determine whether a log line should
// be promoted. By default, the value is "error".
func ErrorKey(value string) Option {
	return func(l *Levels) { l.errorKey = value }
}

// PromoteErrorToLevel sets the log level that log lines containing errors
// should be promoted to. By default, the value is levels.Error.
func PromoteErrorToLevel(value LogLevel) Option {
	return func(l *Levels) { l.promoteErrorToLevel = value }
}
