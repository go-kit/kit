package server

import (
	"golang.org/x/net/context"
)

// Gate wraps an endpoint with a gating function. If the gating function
// returns an error, the request is aborted, and that error is returned.
func Gate(allow func(context.Context, Request) error, next Endpoint) Endpoint {
	return func(ctx context.Context, req Request) (Response, error) {
		if err := allow(ctx, req); err != nil {
			return nil, err
		}
		return next(ctx, req)
	}
}
