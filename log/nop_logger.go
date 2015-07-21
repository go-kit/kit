package log

type nopLogger struct{}

func (nopLogger) Log(...interface{}) error { return nil }

func NewNopLogger() Logger { return nopLogger{} }
