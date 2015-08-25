package http

import (
	"net/http"

	"golang.org/x/net/context"
)

// RequestFunc may take information from an HTTP request and put it into a
// request context. In Servers, BeforeFuncs are executed prior to invoking the
// endpoint. In Clients, BeforeFuncs are executed after creating the request
// but prior to invoking the HTTP client.
type RequestFunc func(context.Context, *http.Request) context.Context

// ResponseFunc may take information from a request context and use it to
// manipulate a ResponseWriter. ResponseFuncs are only executed in servers,
// after invoking the endpoint but prior to writing a response.
type ResponseFunc func(context.Context, http.ResponseWriter)

// SetContentType returns a ResponseFunc that sets the Content-Type header to
// the provided value.
func SetContentType(contentType string) ResponseFunc {
	return SetResponseHeader("Content-Type", contentType)
}

// SetResponseHeader returns a ResponseFunc that sets the specified header.
func SetResponseHeader(key, val string) ResponseFunc {
	return func(_ context.Context, w http.ResponseWriter) {
		w.Header().Set(key, val)
	}
}

// SetRequestHeader returns a RequestFunc that sets the specified header.
func SetRequestHeader(key, val string) RequestFunc {
	return func(ctx context.Context, r *http.Request) context.Context {
		r.Header.Set(key, val)
		return ctx
	}
}
