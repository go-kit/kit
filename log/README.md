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

Output:
```
question="what is the meaning of life?" answer=42
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

Output:
```
txid=12345 query="some=query"
txid=12345 query="some=query" answer=42
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
	
	stdlog.Print("I sure like pie")
}
```

Output
```
{"msg":"I sure like pie","ts":"2016/01/28 19:41:08"}
```

Adding a timestamp to contextual logs

```go
func handle(logger log.Logger, req *Request) {
	ctx := log.NewContext(logger).With("ts", log.DefaultTimestampUTC, "query", req.Query)
	ctx.Log()

	answer, err := process(ctx, req.Query)
	if err != nil {
		ctx.Log("err", err)
		return
	}

	ctx.Log("answer", answer)
}
```

Output
```
ts=2016-01-29T00:46:04Z query="some=query"
ts=2016-01-29T00:46:04Z query="some=query" answer=42
```

Adding caller info to contextual logs

```go
func handle(logger log.Logger, req *Request) {
	ctx := log.NewContext(logger).With("caller", log.DefaultCaller, "query", req.Query)
	ctx.Log()

	answer, err := process(ctx, req.Query)
	if err != nil {
		ctx.Log("err", err)
		return
	}

	ctx.Log("answer", answer)
}
```

Output
```
caller=logger.go:20 query="some=query"
caller=logger.go:28 query="some=query" answer=42
```
