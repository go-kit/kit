package main

import (
	"time"

	"github.com/go-kit/kit/examples/addsvc/server"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
)

type pureAddService struct{}

func (pureAddService) Sum(a, b int) int { return a + b }

func (pureAddService) Concat(a, b string) string { return a + b }

type loggingMiddleware struct {
	server.AddService
	log.Logger
}

func (m loggingMiddleware) Sum(a, b int) (v int) {
	defer func(begin time.Time) {
		m.Logger.Log(
			"method", "sum",
			"a", a,
			"b", b,
			"v", v,
			"took", time.Since(begin),
		)
	}(time.Now())
	v = m.AddService.Sum(a, b)
	return
}

func (m loggingMiddleware) Concat(a, b string) (v string) {
	defer func(begin time.Time) {
		m.Logger.Log(
			"method", "concat",
			"a", a,
			"b", b,
			"v", v,
			"took", time.Since(begin),
		)
	}(time.Now())
	v = m.AddService.Concat(a, b)
	return
}

type instrumentingMiddleware struct {
	server.AddService
	requestDuration metrics.TimeHistogram
}

func (m instrumentingMiddleware) Sum(a, b int) (v int) {
	defer func(begin time.Time) {
		methodField := metrics.Field{Key: "method", Value: "sum"}
		m.requestDuration.With(methodField).Observe(time.Since(begin))
	}(time.Now())
	v = m.AddService.Sum(a, b)
	return
}

func (m instrumentingMiddleware) Concat(a, b string) (v string) {
	defer func(begin time.Time) {
		methodField := metrics.Field{Key: "method", Value: "concat"}
		m.requestDuration.With(methodField).Observe(time.Since(begin))
	}(time.Now())
	v = m.AddService.Concat(a, b)
	return
}
