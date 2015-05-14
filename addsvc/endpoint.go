package main

import (
	"golang.org/x/net/context"

	"github.com/go-kit/kit/server"
)

// makeEndpoint returns a server.Endpoint wrapping the passed Add. If Add were
// an interface with multiple methods, we'd need individual endpoints for
// each.
//
// This function is just boiler-plate; in theory, it could be generated.
func makeEndpoint(a Add) server.Endpoint {
	return func(ctx context.Context, req server.Request) (server.Response, error) {
		select {
		case <-ctx.Done():
			return nil, server.ErrContextCanceled
		default:
		}

		addReq, ok := req.(*request)
		if !ok {
			return nil, server.ErrBadCast
		}

		v := a(addReq.A, addReq.B)

		return response{
			V: v,
		}, nil
	}
}
