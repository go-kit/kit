package twirp

import (
	"context"
	"github.com/twitchtv/twirp"
	"net/http"
)

// ClientRequestFunc may modify the context. ClientRequestFuncs are executed
// after creating the request but prior to sending the Twirp request to
// the server.
type ClientRequestFunc func(context.Context) (context.Context, error)

// ServerRequestFunc may take information from the context. ServerRequestFuncs are
// executed prior to invoking the endpoint.
type ServerRequestFunc func(context.Context) context.Context

// ServerResponseFunc may modify the context. ServerResponseFuncs are only executed in
// servers, after invoking the endpoint but prior to writing a response.
type ServerResponseFunc func(context.Context) (context.Context, error)

// ClientResponseFunc may take information from the context. ClientResponseFuncs are only executed in
// clients, after a request has been made, but prior to it being decoded.
type ClientResponseFunc func(context.Context) (context.Context, error)

// SetResponseHeader returns a ServerResponseFunc that sets the given header.
func SetResponseHeader(key, val string) ServerResponseFunc {
	return func(ctx context.Context) (context.Context, error) {
		err := twirp.SetHTTPResponseHeader(ctx, key, val)
		return ctx, err
	}
}

// SetRequestHeader returns a RequestFunc that sets the given header.
func SetRequestHeader(key, val string) ClientRequestFunc {
	h := &http.Header{}
	h.Set(key, val)
	return func(ctx context.Context) (context.Context, error) {
		return twirp.WithHTTPRequestHeaders(ctx, *h)
	}
}
