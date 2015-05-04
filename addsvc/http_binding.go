package main

import (
	"net/http"
	"time"

	"github.com/peterbourgon/gokit/metrics"
)

// HTTP bindings require no service-specific declarations, and so are defined
// in transport/http.

func httpInstrument(requests metrics.Counter, duration metrics.Histogram) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			defer func(begin time.Time) { duration.Observe(time.Since(begin).Nanoseconds()) }(time.Now())
			next.ServeHTTP(w, r)
		})
	}
}
