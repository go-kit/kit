# package transport/httprp

`package transport/httprp` provides an HTTP reverse-proxy transport.

## Rationale

HTTP server applications often associate multiple handlers with a single HTTP listener, each handler differentiated by the request URI and/or HTTP method.  Handlers that perform business-logic in the app can implement the `Endpoint` interface and be exposed using the `package transport/http` server.  Handlers that need to proxy the request to another HTTP endpoint can do so with this package by simply specifying the base URL to forward the request to.

## Usage

The following example uses the [Gorilla Mux](https://github.com/gorilla/mux) router to illustrate how a mixture of proxying and non-proxying request handlers can be used with a single listener:

```go
import (
	"net/http"
	"net/url"

	kithttp "github.com/go-kit/kit/transport/http"
	kithttprp "github.com/go-kit/kit/transport/httprp"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
)

func main() {
	router := mux.NewRouter()

	// server HTTP endpoint handled here
	router.Handle("/foo",
		kithttp.NewServer(
			context.Background(),
			func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil },
			func(*http.Request) (interface{}, error) { return struct{}{}, nil },
			func(http.ResponseWriter, interface{}) error { return nil },
		)).Methods("GET")

	// proxy endpoint, forwards requests to http://other.service.local/base/bar
	remoteServiceURL, _ := url.Parse("http://other.service.local/base")
	router.Handle("/bar",
		kithttprp.NewServer(
			context.Background(),
			remoteServiceURL,
		)).Methods("GET")

	http.ListenAndServe(":8080", router)
}
```

You can also supply a set of `RequestFunc` functions to be run before proxying the request.  This can be useful for adding request headers required by the backend system (e.g. API tokens).
