package level

import (
	"github.com/go-kit/kit/log"
)

var (
	global = AllowingNone()
	/*
	   Alternately:
	   global atomic.Value
	*/
)

/* Alternately:
func init() {
	global.Store(errorOnly{})
}
*/

func resetGlobal(proposed Leveler) {
	global = proposed
	/* Alternately:
	global.Store(proposed)
	*/
}

func AllowAll() {
	AllowDebugAndAbove()
}

func AllowDebugAndAbove() {
	resetGlobal(AllowingDebugAndAbove())
}

func AllowInfoAndAbove() {
	resetGlobal(AllowingInfoAndAbove())
}

func AllowWarnAndAbove() {
	resetGlobal(AllowingWarnAndAbove())
}

func AllowErrorOnly() {
	resetGlobal(AllowingErrorOnly())
}

func AllowNone() {
	resetGlobal(AllowingNone())
}

func getGlobal() Leveler {
	return global
	/* Alternately:
	return global.Load().(Leveler)
	*/
}

func Debug(logger log.Logger) log.Logger {
	return getGlobal().Debug(logger)
}

func Info(logger log.Logger) log.Logger {
	return getGlobal().Info(logger)
}

func Warn(logger log.Logger) log.Logger {
	return getGlobal().Warn(logger)
}

func Error(logger log.Logger) log.Logger {
	return getGlobal().Error(logger)
}
