package endpoint

import (
	"errors"

	"golang.org/x/net/context"
)

// Endpoint is the fundamental building block of servers and clients.
// It represents a single RPC method.
type Endpoint func(ctx context.Context, request interface{}) (response interface{}, err error)

// Middleware is a chainable behavior modifier for endpoints.
type Middleware func(Endpoint) Endpoint

// ErrBadCast indicates an unexpected concrete request or response struct was
// received from an endpoint.
var ErrBadCast = errors.New("bad cast")

// ErrContextCanceled indicates the request context was canceled.
var ErrContextCanceled = errors.New("context canceled")

// Chain is a helper function for composing middlewares. Requests will
// traverse them in the order they're declared. That is, the first middleware
// is treated as the outermost middleware.
func Chain(outer Middleware, others ...Middleware) Middleware {
	return func(next Endpoint) Endpoint {
		for i := len(others) - 1; i >= 0; i-- { // reverse
			next = others[i](next)
		}
		return outer(next)
	}
}
