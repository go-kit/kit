package log_test

import (
	"net/url"
	"os"

	"github.com/go-kit/kit/log"
)

func Example_stdout() {
	w := log.NewSyncWriter(os.Stdout)
	logger := log.NewLogfmtLogger(w)

	reqUrl := &url.URL{
		Scheme: "https",
		Host:   "github.com",
		Path:   "/go-kit/kit",
	}

	logger.Log("method", "GET", "url", reqUrl)

	// Output:
	// method=GET url=https://github.com/go-kit/kit
}

func ExampleContext() {
	w := log.NewSyncWriter(os.Stdout)
	logger := log.NewLogfmtLogger(w)
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
