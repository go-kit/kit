package log

type Levels interface {
	With(...interface{}) Levels
	Debug(...interface{}) error
	Info(...interface{}) error
	Warn(...interface{}) error
	Error(...interface{}) error
	Crit(...interface{}) error
}

func NewLevels(logger Logger, keyvals ...interface{}) Levels {
	return levels(NewContext(logger, keyvals...))
}

type levels Context

func (l levels) With(keyvals ...interface{}) Levels {
	return levels(Context(l).With(keyvals...))
}

func (l levels) Debug(keyvals ...interface{}) error {
	return NewContext(l.logger).With("level", "debug").With(l.keyvals...).Log(keyvals...)
}

func (l levels) Info(keyvals ...interface{}) error {
	return NewContext(l.logger).With("level", "info").With(l.keyvals...).Log(keyvals...)
}

func (l levels) Warn(keyvals ...interface{}) error {
	return NewContext(l.logger).With("level", "warn").With(l.keyvals...).Log(keyvals...)
}

func (l levels) Error(keyvals ...interface{}) error {
	return NewContext(l.logger).With("level", "error").With(l.keyvals...).Log(keyvals...)
}

func (l levels) Crit(keyvals ...interface{}) error {
	return NewContext(l.logger).With("level", "crit").With(l.keyvals...).Log(keyvals...)
}
