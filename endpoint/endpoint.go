package endpoint

import (
	"context"
)

// Endpoint is the fundamental building block of servers and clients.
// It represents a single RPC method.
type Endpoint func(ctx context.Context, request interface{}) (response interface{}, err error)

// Nop is an endpoint that does nothing and returns a nil error.
// Useful for tests.
func Nop(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil }

// Middleware is a chainable behavior modifier for endpoints.
type Middleware func(Endpoint) Endpoint

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

// Failer is an interface that should be implemented by response types that
// hold error properties as to separate business errors from transport errors.
// If the response type can hold business errors it is highly advised to
// implement Failer.
// Response encoders can check if responses are Failer, and if so if they've
// failed encode them using a separate write path based on the error.
// Endpoint middlewares can test if a response type failed and also act or
// report upon it.
//
// The addsvc example shows Failer's intended usage.
type Failer interface {
	Failed() error
}
