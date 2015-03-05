package main

import (
	"github.com/peterbourgon/gokit/server"
	"golang.org/x/net/context"
)

func makeEndpoint(a Add) server.Endpoint {
	return func(ctx context.Context, req server.Request) (server.Response, error) {
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
