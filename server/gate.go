package server

import (
	"golang.org/x/net/context"
)

// Gate returns a middleware that gates requests. If the gating function
// returns an error, the request is aborted, and that error is returned.
func Gate(allow func(context.Context, Request) error) func(Endpoint) Endpoint {
	return func(next Endpoint) Endpoint {
		return func(ctx context.Context, req Request) (Response, error) {
			if err := allow(ctx, req); err != nil {
				return nil, err
			}
			return next(ctx, req)
		}
	}
}
