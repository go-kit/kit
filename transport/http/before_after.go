package http

import (
	"net/http"

	"golang.org/x/net/context"
)

// BeforeFunc may take information from a HTTP request and put it into a
// request context. BeforeFuncs are executed in HTTP bindings, prior to
// invoking the endpoint.
type BeforeFunc func(context.Context, *http.Request) context.Context

// AfterFunc may take information from a request context and use it to
// manipulate a ResponseWriter. AfterFuncs are executed in HTTP bindings,
// after invoking the endpoint but prior to writing a response.
type AfterFunc func(context.Context, http.ResponseWriter)

// SetContentType returns an AfterFunc that sets the HTTP Content-Type header
// to the provided value.
func SetContentType(value string) AfterFunc {
	return func(_ context.Context, w http.ResponseWriter) {
		w.Header().Set("Content-Type", value)
	}
}
