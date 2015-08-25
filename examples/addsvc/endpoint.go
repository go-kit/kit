package main

import (
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/examples/addsvc/reqrep"
)

// makeEndpoint returns an endpoint wrapping the passed Add. If Add were an
// interface with multiple methods, we'd need individual endpoints for each.
//
// This function is just boiler-plate; in theory, it could be generated.
func makeEndpoint(a Add) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		select {
		default:
		case <-ctx.Done():
			return nil, endpoint.ErrContextCanceled
		}

		addReq, ok := request.(reqrep.AddRequest)
		if !ok {
			return nil, endpoint.ErrBadCast
		}

		v := a(ctx, addReq.A, addReq.B)
		return reqrep.AddResponse{V: v}, nil
	}
}
