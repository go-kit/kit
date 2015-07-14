package levels

import "github.com/go-kit/kit/log"

// Levels provides a leveled logging wrapper around a logger. It has five
// levels: debug, info, warning (warn), error, and critical (crit). If you
// want a different set of levels, you can create your own levels type very
// easily, and you can elide the configuration.
type Levels struct {
	ctx log.Context
	cfg *config
}

// New creates a new leveled logger, wrapping the passed logger.
func New(logger log.Logger, options ...Option) Levels {
	cfg := &config{
		levelKey:   "level",
		debugValue: "debug",
		infoValue:  "info",
		warnValue:  "warn",
		errorValue: "error",
		critValue:  "crit",
	}
	for _, option := range options {
		option(cfg)
	}
	return Levels{
		ctx: log.NewContext(logger),
		cfg: cfg,
	}
}

// With returns a new leveled logger that includes keyvals in all log events.
func (l Levels) With(keyvals ...interface{}) Levels {
	return Levels{
		ctx: l.ctx.With(keyvals...),
		cfg: l.cfg,
	}
}

// Debug logs a debug event along with keyvals.
func (l Levels) Debug(keyvals ...interface{}) error {
	return l.ctx.WithPrefix(l.cfg.levelKey, l.cfg.debugValue).Log(keyvals...)
}

// Info logs an info event along with keyvals.
func (l Levels) Info(keyvals ...interface{}) error {
	return l.ctx.WithPrefix(l.cfg.levelKey, l.cfg.infoValue).Log(keyvals...)
}

// Warn logs a warn event along with keyvals.
func (l Levels) Warn(keyvals ...interface{}) error {
	return l.ctx.WithPrefix(l.cfg.levelKey, l.cfg.warnValue).Log(keyvals...)
}

// Error logs an error event along with keyvals.
func (l Levels) Error(keyvals ...interface{}) error {
	return l.ctx.WithPrefix(l.cfg.levelKey, l.cfg.errorValue).Log(keyvals...)
}

// Crit logs a crit event along with keyvals.
func (l Levels) Crit(keyvals ...interface{}) error {
	return l.ctx.WithPrefix(l.cfg.levelKey, l.cfg.critValue).Log(keyvals...)
}

type config struct {
	levelKey string

	debugValue string
	infoValue  string
	warnValue  string
	errorValue string
	critValue  string
}

// Option sets a parameter for leveled loggers.
type Option func(*config)

// Key sets the key for the field used to indicate log level. By default,
// the key is "level".
func Key(key string) Option {
	return func(c *config) { c.levelKey = key }
}

// DebugValue sets the value for the field used to indicate the debug log
// level. By default, the value is "debug".
func DebugValue(value string) Option {
	return func(c *config) { c.debugValue = value }
}

// InfoValue sets the value for the field used to indicate the info log level.
// By default, the value is "info".
func InfoValue(value string) Option {
	return func(c *config) { c.infoValue = value }
}

// WarnValue sets the value for the field used to indicate the warning log
// level. By default, the value is "warn".
func WarnValue(value string) Option {
	return func(c *config) { c.warnValue = value }
}

// ErrorValue sets the value for the field used to indicate the error log
// level. By default, the value is "error".
func ErrorValue(value string) Option {
	return func(c *config) { c.errorValue = value }
}

// CritValue sets the value for the field used to indicate the critical log
// level. By default, the value is "crit".
func CritValue(value string) Option {
	return func(c *config) { c.critValue = value }
}
