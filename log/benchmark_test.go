package log_test

import (
	"testing"

	"github.com/peterbourgon/gokit/log"
)

func benchmarkRunner(b *testing.B, logger log.Logger, f func(log.Logger)) {
	logger = logger.With("common_key", "common_value")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f(logger)
	}
}

var (
	simpleMessage     = func(logger log.Logger) { logger.Log("foo") }
	contextualMessage = func(logger log.Logger) { logger.Log("bar", "foo_key", "foo_value") }
)
