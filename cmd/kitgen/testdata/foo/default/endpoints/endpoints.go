package endpoints

import "context"

import "github.com/go-kit/kit/endpoint"

type BarRequest struct {
	I int
	S string
}
type BarResponse struct {
	S   string
	Err error
}

func makeBarEndpoint(f stubFooService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(BarRequest)
		s, err := f.Bar(ctx, req.I, req.S)
		return BarResponse{S: s, Err: err}, nil
	}
}

type Endpoints struct {
	Bar endpoint.Endpoint
}