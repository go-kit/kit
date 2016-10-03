package level

import (
	"github.com/go-kit/kit/log"
)

var (
	// Alternately, we could use a similarly inert logger that does nothing but
	// return a given error value.
	nop = log.NewNopLogger()
)

type Leveler interface {
	Debug(logger log.Logger) log.Logger
	Info(logger log.Logger) log.Logger
	Warn(logger log.Logger) log.Logger
	Error(logger log.Logger) log.Logger
}

func withLevel(level string, logger log.Logger) log.Logger {
	return log.NewContext(logger).With("level", level)
}

type debugAndAbove struct{}

func (debugAndAbove) Debug(logger log.Logger) log.Logger {
	return withLevel("debug", logger)
}

func (debugAndAbove) Info(logger log.Logger) log.Logger {
	return withLevel("info", logger)
}

func (debugAndAbove) Warn(logger log.Logger) log.Logger {
	return withLevel("warn", logger)
}

func (debugAndAbove) Error(logger log.Logger) log.Logger {
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

func AllowingAll() Leveler {
	return AllowingDebugAndAbove()
}

func AllowingDebugAndAbove() Leveler {
	return debugAndAbove{}
}

func AllowingInfoAndAbove() Leveler {
	return infoAndAbove{}
}

func AllowingWarnAndAbove() Leveler {
	return warnAndAbove{}
}

func AllowingErrorOnly() Leveler {
	return errorOnly{}
}

func AllowingNone() Leveler {
	return none{}
}
