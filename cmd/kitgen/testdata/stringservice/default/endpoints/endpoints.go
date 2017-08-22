package endpoints

import "golang.org/x/net/context"
import "context"

import "github.com/go-kit/kit/endpoint"

type ConcatRequest struct {
	A string
	B string
}
type ConcatResponse struct {
	S   string
	Err error
}

func makeConcatEndpoint(s stubService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ConcatRequest)
		string1, err := s.Concat(ctx, req.A, req.B)
		return ConcatResponse{S: string1, Err: err}, nil
	}
}

type CountRequest struct {
	S string
}
type CountResponse struct {
	Count int
}

func makeCountEndpoint(s stubService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CountRequest)
		count := s.Count(ctx, req.S)
		return CountResponse{Count: count}, nil
	}
}

type Endpoints struct {
	Concat endpoint.Endpoint
	Count  endpoint.Endpoint
}
