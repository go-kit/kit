package awslambda

import (
	"context"
)

// ServerRequestFunc may take information from the received
// payload and use it to place items in the request scoped context.
// ServerRequestFuncs are executed prior to invoking the endpoint and
// decoding of the payload.
type ServerRequestFunc func(ctx context.Context, payload []byte) context.Context

// ServerResponseFunc may take information from a request context
// and use it to manipulate response (before marshalled.)
// ServerResponseFunc are only executed after invoking the endpoint
// but prior to returning a response.
type ServerResponseFunc func(ctx context.Context, response interface{}) context.Context

// ServerFinalizerFunc is executed at the end of Invocation.
// This can be used for logging purposes.
type ServerFinalizerFunc func(ctx context.Context, resp []byte, err error)
