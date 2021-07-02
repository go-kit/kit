package twirp

import (
	"context"
	"net/http"
)

// ClientRequestFunc may modify the context. ClientRequestFuncs are executed
// after creating the request but prior to sending the Twirp request to
// the server.
type ClientRequestFunc func(context.Context, *http.Header) context.Context

// ServerRequestFunc may take information from the context. ServerRequestFuncs are
// executed prior to invoking the endpoint.
type ServerRequestFunc func(context.Context, http.Header) context.Context

// ServerResponseFunc may modify the context. ServerResponseFuncs are only executed in
// servers, after invoking the endpoint but prior to writing a response.
type ServerResponseFunc func(context.Context) context.Context

// ClientResponseFunc may take information from the context. ClientResponseFuncs are only executed in
// clients, after a request has been made, but prior to it being decoded.
type ClientResponseFunc func(context.Context) context.Context

// SetRequestHeader returns a RequestFunc that sets the given header. It uses the standard net/http/header Add function and will append the specified value if others already exist.
func SetRequestHeader(key, val string) ClientRequestFunc {
	return func(ctx context.Context, header *http.Header) context.Context {
		header.Add(key, val)
		return ctx
	}
}
