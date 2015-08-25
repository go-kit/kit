package main

import (
	"fmt"
	"time"

	"github.com/go-kit/kit/metrics"
)

func instrumentingMiddleware(
	requestCount metrics.Counter,
	requestLatency metrics.TimeHistogram,
	countResult metrics.Histogram,
) ServiceMiddleware {
	return func(next StringService) StringService {
		return instrmw{requestCount, requestLatency, countResult, next}
	}
}

type instrmw struct {
	requestCount   metrics.Counter
	requestLatency metrics.TimeHistogram
	countResult    metrics.Histogram
	StringService
}

func (mw instrmw) Uppercase(s string) (output string, err error) {
	defer func(begin time.Time) {
		methodField := metrics.Field{Key: "method", Value: "uppercase"}
		errorField := metrics.Field{Key: "error", Value: fmt.Sprintf("%v", err)}
		mw.requestCount.With(methodField).With(errorField).Add(1)
		mw.requestLatency.With(methodField).With(errorField).Observe(time.Since(begin))
	}(time.Now())

	output, err = mw.StringService.Uppercase(s)
	return
}

func (mw instrmw) Count(s string) (n int) {
	defer func(begin time.Time) {
		methodField := metrics.Field{Key: "method", Value: "count"}
		errorField := metrics.Field{Key: "error", Value: fmt.Sprintf("%v", error(nil))}
		mw.requestCount.With(methodField).With(errorField).Add(1)
		mw.requestLatency.With(methodField).With(errorField).Observe(time.Since(begin))
		mw.countResult.Observe(int64(n))
	}(time.Now())

	n = mw.StringService.Count(s)
	return
}
