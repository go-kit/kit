package service

import (
	"errors"

	"github.com/go-kit/kit/endpoint"
)

// Service represents a collection of endpoints (i.e. methods). It may be one
// instance of a microservice, or it may abstract over multiple identical
// instances of the same microservice.
type Service interface {
	Endpoint(method string) (endpoint.Endpoint, error)
}

// The Func type is an adapter to allow the use of ordinary functions as
// services. If f is a function with the appropriate signature, Func(f) is a
// Service that calls f.
type Func func(method string) (endpoint.Endpoint, error)

// Endpoint calls f(method).
func (f Func) Endpoint(method string) (endpoint.Endpoint, error) {
	return f(method)
}

// ErrMethodNotFound indicates a method has no corresponding endpoint.
var ErrMethodNotFound = errors.New("method not found")

// Middleware is a chainable behavior modifier for services.
type Middleware func(Service) Service

// Chain is a helper function for composing middlewares. Requests will
// traverse them in the order they're declared. That is, the first middleware
// is treated as the outermost middleware.
func Chain(outer Middleware, others ...Middleware) Middleware {
	return func(next Service) Service {
		for i := len(others) - 1; i >= 0; i-- { // reverse
			next = others[i](next)
		}
		return outer(next)
	}
}
