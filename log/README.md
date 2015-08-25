# package log

`package log` provides a minimal interface for structured logging in services.
It may be wrapped to encode conventions, enforce type-safety, etc.
It can be used for both typical application log events, and log-structured data streams.

## Rationale

TODO

## Usage

Typical application logging.

```go
import (
  "os"

  "github.com/go-kit/kit/log"
)

func main() {
	logger := log.NewLogfmtLogger(os.Stderr)
	logger.Log("question", "what is the meaning of life?", "answer", 42)
}
```

Contextual logging.

```go
func handle(logger log.Logger, req *Request) {
	ctx := log.NewContext(logger).With("txid", req.TransactionID, "query", req.Query)
	ctx.Log()

	answer, err := process(ctx, req.Query)
	if err != nil {
		ctx.Log("err", err)
		return
	}

	ctx.Log("answer", answer)
}
```

Redirect stdlib log to gokit logger.

```go
import (
	"os"
	stdlog "log"
	kitlog "github.com/go-kit/kit/log"
)

func main() {
	logger := kitlog.NewJSONLogger(os.Stdout)
	stdlog.SetOutput(kitlog.NewStdlibAdapter(logger))
}
```
