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

// ContextCanceled indicates the request context was canceled.
var ErrContextCanceled = errors.New("context canceled")
