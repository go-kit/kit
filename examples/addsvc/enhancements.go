package main

import (
	"time"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
)

func logging(logger log.Logger) func(Add) Add {
	return func(next Add) Add {
		return func(ctx context.Context, a, b int64) (v int64) {
			defer func(begin time.Time) {
				logger.Log("a", a, "b", b, "result", v, "took", time.Since(begin))
			}(time.Now())
			v = next(ctx, a, b)
			return
		}
	}
}

func instrument(requests metrics.Counter, duration metrics.TimeHistogram) func(Add) Add {
	return func(next Add) Add {
		return func(ctx context.Context, a, b int64) int64 {
			defer func(begin time.Time) {
				requests.Add(1)
				duration.Observe(time.Since(begin))
			}(time.Now())
			return next(ctx, a, b)
		}
	}
}
