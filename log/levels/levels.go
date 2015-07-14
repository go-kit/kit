package levels

import "github.com/go-kit/kit/log"

// Levels provides leveled logging.
type Levels struct {
	ctx  log.Context
	opts *config
}

// New creates a new leveled logger.
func New(logger log.Logger, options ...Option) Levels {
	opts := &config{
		levelKey:   "level",
		debugValue: "debug",
		infoValue:  "info",
		warnValue:  "warn",
		errorValue: "error",
		critValue:  "crit",
	}
	for _, option := range options {
		option(opts)
	}
	return Levels{
		ctx:  log.NewContext(logger),
		opts: opts,
	}
}

// With returns a new leveled logger that includes keyvals in all log events.
func (l Levels) With(keyvals ...interface{}) Levels {
	return Levels{
		ctx:  l.ctx.With(keyvals...),
		opts: l.opts,
	}
}

// Debug logs a debug event along with keyvals.
func (l Levels) Debug(keyvals ...interface{}) error {
	return l.ctx.WithPrefix(l.opts.levelKey, l.opts.debugValue).Log(keyvals...)
}

// Info logs an info event along with keyvals.
func (l Levels) Info(keyvals ...interface{}) error {
	return l.ctx.WithPrefix(l.opts.levelKey, l.opts.infoValue).Log(keyvals...)
}

// Warn logs a warn event along with keyvals.
func (l Levels) Warn(keyvals ...interface{}) error {
	return l.ctx.WithPrefix(l.opts.levelKey, l.opts.warnValue).Log(keyvals...)
}

// Error logs an error event along with keyvals.
func (l Levels) Error(keyvals ...interface{}) error {
	return l.ctx.WithPrefix(l.opts.levelKey, l.opts.errorValue).Log(keyvals...)
}

// Crit logs a crit event along with keyvals.
func (l Levels) Crit(keyvals ...interface{}) error {
	return l.ctx.WithPrefix(l.opts.levelKey, l.opts.critValue).Log(keyvals...)
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
	return func(o *config) { o.levelKey = key }
}

// DebugValue sets the value for the field used to indicate the debug log
// level. By default, the value is "debug".
func DebugValue(value string) Option {
	return func(o *config) { o.debugValue = value }
}

// InfoValue sets the value for the field used to indicate the debug log
// level. By default, the value is "info".
func InfoValue(value string) Option {
	return func(o *config) { o.infoValue = value }
}

// WarnValue sets the value for the field used to indicate the debug log
// level. By default, the value is "debug".
func WarnValue(value string) Option {
	return func(o *config) { o.warnValue = value }
}

// ErrorValue sets the value for the field used to indicate the debug log
// level. By default, the value is "error".
func ErrorValue(value string) Option {
	return func(o *config) { o.errorValue = value }
}

// CritValue sets the value for the field used to indicate the debug log
// level. By default, the value is "debug".
func CritValue(value string) Option {
	return func(o *config) { o.critValue = value }
}
