# package log

`package log` provides an interface for structured logging.

It may be wrapped to encode conventions, enforce type-safety, etc.
It can be used for both typical application log events, and log-structured data streams.

## Rationale

TODO

## Usage

Typical application logging.

```go
import "github.com/go-kit/kit/log"

func main() {
	logger := log.NewPrefixLogger(os.Stderr)
	logger.Log("question", "what is the meaning of life?", "answer", 42)
}
```

The log lines are printed in tuples separated with by a `=` symbol.

```
question=what is the meaning of life? answer=42
```

Contextual logging.

```go
func handle(logger log.Logger, req *Request) {
	logger = log.With(logger, "txid", req.TransactionID, "query", req.Query)
	logger.Log()

	answer, err := process(logger, req.Query)
	if err != nil {
		logger.Log("err", err)
		return
	}

	logger.Log("answer", answer)
}
```

Any log lines produced from the contextual logger will have `txid` and `query`
prepended.

```
txid=0EBC0A62-B042-4974-8517-2CDAC1D4631E query=what is the meaning of life? answer=42
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
