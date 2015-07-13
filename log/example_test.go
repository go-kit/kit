package log_test

import (
	"os"

	"github.com/go-kit/kit/log"
)

func ExampleContext() {
	logger := log.NewLogfmtLogger(os.Stdout)
	logger.Log("foo", 123)
	ctx := log.NewContext(logger).With("level", "info")
	ctx.Log()
	ctx = ctx.With("msg", "hello")
	ctx.Log()
	ctx.With("a", 1).Log("b", 2)

	// Output:
	// foo=123
	// level=info
	// level=info msg=hello
	// level=info msg=hello a=1 b=2
}
