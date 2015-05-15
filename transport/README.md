# package transport

`package transport` defines interfaces for service transports and codecs.
It also provides implementations for transports and codecs that aren't already well-defined by other packages.
The most common use case for `package transport` is probably to bind a gokit [server.Endpoint][] with a stdlib [http.Handler][], via gokit's [http.Binding][].
Refer to the [addsvc][] example service to see how to make [Thrift][] or [gRPC][] transport bindings.

[server.Endpoint]: https://godoc.org/github.com/go-kit/kit/server/#Endpoint
[http.Handler]: https://golang.org/pkg/net/http/#Handler
[http.Binding]: https://godoc.org/github.com/go-kit/kit/transport/http/#Binding
[addsvc]: https://github.com/go-kit/kit/blob/319c1c7129a146b541bbbaf18e2502bf32c603c5/addsvc/main.go
[Thrift]: https://github.com/go-kit/kit/blob/319c1c7129a146b541bbbaf18e2502bf32c603c5/addsvc/main.go#L142-192
[gRPC]: https://github.com/go-kit/kit/blob/319c1c7129a146b541bbbaf18e2502bf32c603c5/addsvc/main.go#L102-119

## Rationale

TODO

## Usage

Bind a gokit [server.Endpoint][] with a stdlib [http.Handler][].

```go
import (
	"net/http"
	"reflect"

	"golang.org/x/net/context"

	jsoncodec "github.com/go-kit/kit/transport/codec/json"
	httptransport "github.com/go-kit/kit/transport/http"
)

type request struct{}

func main() {
	var (
		ctx         = context.Background()
		requestType = reflect.TypeOf(request{})
		codec       = jsoncodec.New()
		e           = makeEndpoint()
		before      = []httptransport.BeforeFunc{}
		after       = []httptransport.AfterFunc{}
	)
	handler := httptransport.NewBinding(ctx, requestType, codec, e, before, after)
	mux := http.NewServeMux()
	mux.Handle("/path", handler)
	http.ListenAndServe(":8080", mux)
}
```
