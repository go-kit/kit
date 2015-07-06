package http

import (
	"net/http"

	"golang.org/x/net/context"
)

// BeforeFunc may take information from an HTTP request and put it into a
// request context. BeforeFuncs are executed in the handler, prior to invoking
// the endpoint.
type BeforeFunc func(context.Context, *http.Request) context.Context

// AfterFunc may take information from a request context and use it to
// manipulate a ResponseWriter. AfterFuncs are executed in the handler, after
// invoking the endpoint but prior to writing a response.
type AfterFunc func(context.Context, http.ResponseWriter)

// SetContentType returns an AfterFunc that sets the Content-Type header to
// the provided value.
func SetContentType(contentType string) AfterFunc {
	return SetHeader("Content-Type", contentType)
}

// SetHeader returns an AfterFunc that sets the specified header on the
// response.
func SetHeader(key, val string) AfterFunc {
	return func(_ context.Context, w http.ResponseWriter) {
		w.Header().Set(key, val)
	}
}
