package main

import (
	"time"

	"github.com/peterbourgon/gokit/log"
)

func logging(logger log.Logger, add Add) Add {
	return func(a, b int64) (v int64) {
		defer func(begin time.Time) {
			logger.Log("a", a, "b", b, "result", v, "took", time.Since(begin))
		}(time.Now())
		v = add(a, b)
		return
	}
}
