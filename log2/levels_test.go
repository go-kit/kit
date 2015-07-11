package log_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/go-kit/kit/log2"
)

func ExampleLevels() {
	logger := log.NewLevels(log.NewLogfmtLogger(os.Stdout))
	logger.Debug("msg", "hello")
	logger.With("context", "foo").Warn("err", "error")
	// Output:
	// level=debug msg=hello
	// level=warn context=foo err=error
}

func BenchmarkLevels(b *testing.B) {
	logger := log.NewLevels(log.NewLogfmtLogger(ioutil.Discard)).With("foo", "bar")
	for i := 0; i < b.N; i++ {
		logger.Debug("key", "val")
	}
}
