package level

import (
	"github.com/go-kit/kit/log"
)

var (
	// Alternately, we could use a similarly inert logger that does nothing but
	// return a given error value.
	nop = log.NewNopLogger()
)

type leveler interface {
	Debug() log.Logger
	Info() log.Logger
	Warn() log.Logger
	Error() log.Logger
}

func withLevel(level string, logger log.Logger) log.Logger {
	return log.NewContext(logger).With("level", level)
}

type debugAndAbove struct {
	log.Logger
}

func (l debugAndAbove) Debug() log.Logger {
	return withLevel("debug", l.Logger)
}

func (l debugAndAbove) Info() log.Logger {
	return withLevel("info", l.Logger)
}

func (l debugAndAbove) Warn() log.Logger {
	return withLevel("warn", l.Logger)
}

func (l debugAndAbove) Error() log.Logger {
	return withLevel("error", l.Logger)
}

type infoAndAbove struct {
	debugAndAbove
}

func (infoAndAbove) Debug() log.Logger {
	return nop
}

type warnAndAbove struct {
	infoAndAbove
}

func (warnAndAbove) Info() log.Logger {
	return nop
}

type errorOnly struct {
	warnAndAbove
}

func (errorOnly) Warn() log.Logger {
	return nop
}

type none struct {
	errorOnly
}

func (none) Error() log.Logger {
	return nop
}

func AllowingAll(logger log.Logger) log.Logger {
	return AllowingDebugAndAbove(logger)
}

func AllowingDebugAndAbove(logger log.Logger) log.Logger {
	if _, ok := logger.(leveler); ok {
		return logger
	}
	return debugAndAbove{logger}
}

func AllowingInfoAndAbove(logger log.Logger) log.Logger {
	switch l := logger.(type) {
	case debugAndAbove:
		return infoAndAbove{l}
	case infoAndAbove, warnAndAbove, errorOnly, none:
		return logger
	default:
		return infoAndAbove{debugAndAbove{logger}}
	}
}

func AllowingWarnAndAbove(logger log.Logger) log.Logger {
	switch l := logger.(type) {
	case debugAndAbove:
		return warnAndAbove{infoAndAbove{l}}
	case infoAndAbove:
		return warnAndAbove{l}
	case warnAndAbove, errorOnly, none:
		return logger
	default:
		return warnAndAbove{infoAndAbove{debugAndAbove{logger}}}
	}
}

func AllowingErrorOnly(logger log.Logger) log.Logger {
	switch l := logger.(type) {
	case debugAndAbove:
		return errorOnly{warnAndAbove{infoAndAbove{l}}}
	case infoAndAbove:
		return errorOnly{warnAndAbove{l}}
	case warnAndAbove:
		return errorOnly{l}
	case errorOnly, none:
		return logger
	default:
		return errorOnly{warnAndAbove{infoAndAbove{debugAndAbove{logger}}}}
	}
}

func AllowingNone(logger log.Logger) log.Logger {
	switch l := logger.(type) {
	case debugAndAbove:
		return none{errorOnly{warnAndAbove{infoAndAbove{l}}}}
	case infoAndAbove:
		return none{errorOnly{warnAndAbove{l}}}
	case warnAndAbove:
		return none{errorOnly{l}}
	case errorOnly:
		return none{l}
	case none:
		return logger
	default:
		return none{errorOnly{warnAndAbove{infoAndAbove{debugAndAbove{logger}}}}}
	}
}

func Debug(logger log.Logger) log.Logger {
	if l, ok := logger.(leveler); ok {
		return l.Debug()
	}
	return nop
}

func Info(logger log.Logger) log.Logger {
	if l, ok := logger.(leveler); ok {
		return l.Info()
	}
	return nop
}

func Warn(logger log.Logger) log.Logger {
	if l, ok := logger.(leveler); ok {
		return l.Warn()
	}
	return nop
}

func Error(logger log.Logger) log.Logger {
	if l, ok := logger.(leveler); ok {
		return l.Error()
	}
	return nop
}
