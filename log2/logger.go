package log

type Logger interface {
	Log(keyvals ...interface{}) error
}

type Context struct {
	logger  Logger
	keyvals []interface{}
}

func NewContext(logger Logger, keyvals ...interface{}) Context {
	if len(keyvals)%2 != 0 {
		panic("bad keyvals")
	}
	return Context{
		logger:  logger,
		keyvals: keyvals,
	}
}

func (c Context) With(keyvals ...interface{}) Context {
	if len(keyvals)%2 != 0 {
		panic("bad keyvals")
	}
	return Context{
		logger:  c.logger,
		keyvals: append(c.keyvals, keyvals...),
	}
}

func (c Context) Log(keyvals ...interface{}) error {
	return c.logger.Log(append(c.keyvals, keyvals...)...)
}
