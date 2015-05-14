package log_test

import (
	"testing"

	"github.com/go-kit/kit/log"
)

func benchmarkRunner(b *testing.B, logger log.Logger, f func(log.Logger)) {
	logger = log.With(logger, "common_key", "common_value")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f(logger)
	}
}

var (
	baseMessage = func(logger log.Logger) { logger.Log("foo_key", "foo_value") }
	withMessage = func(logger log.Logger) { log.With(logger, "a", "b").Log("c", "d") }
)
