package endpoints

import "context"

import "github.com/go-kit/kit/endpoint"

type FooRequest struct {
	I int
}
type FooResponse struct {
	I   int
	Err error
}

func makeFooEndpoint(s stubService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(FooRequest)
		i, err := s.Foo(ctx, req.I)
		return FooResponse{I: i, Err: err}, nil
	}
}

type Endpoints struct {
	Foo endpoint.Endpoint
}
