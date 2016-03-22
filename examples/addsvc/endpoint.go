package main

import (
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/examples/addsvc/server"
)

func makeSumEndpoint(svc server.AddService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*server.SumRequest)
		v := svc.Sum(req.A, req.B)
		return server.SumResponse{V: v}, nil
	}
}

func makeConcatEndpoint(svc server.AddService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*server.ConcatRequest)
		v := svc.Concat(req.A, req.B)
		return server.ConcatResponse{V: v}, nil
	}
}
