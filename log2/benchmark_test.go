package log_test

import (
	"io/ioutil"
	"testing"

	"github.com/go-kit/kit/log2"
)

func BenchmarkContextNoMessage(b *testing.B) {
	logger := log.NewLogfmtLogger(ioutil.Discard)
	ctx := log.NewContext(logger, "module", "benchmark")
	for i := 0; i < b.N; i++ {
		ctx.Log()
	}
}

func BenchmarkContextOneMessage(b *testing.B) {
	logger := log.NewLogfmtLogger(ioutil.Discard)
	ctx := log.NewContext(logger, "module", "benchmark")
	for i := 0; i < b.N; i++ {
		ctx.Log("msg", "hello")
	}
}

func BenchmarkContextWith(b *testing.B) {
	logger := log.NewLogfmtLogger(ioutil.Discard)
	ctx := log.NewContext(logger, "module", "benchmark")
	for i := 0; i < b.N; i++ {
		ctx.With("subcontext", 123).Log("msg", "goodbye")
	}
}
