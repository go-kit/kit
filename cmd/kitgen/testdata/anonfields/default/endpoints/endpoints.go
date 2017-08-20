package endpoints

import "context"
import "encoding/json"
import "errors"
import "net/http"
import "github.com/go-kit/kit/endpoint"
import httptransport "github.com/go-kit/kit/transport/http"

type FooRequest struct {
	I int
	S string
}
type FooResponse struct {
	I   int
	Err error
}

func makeFooEndpoint(s stubService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(FooRequest)
		i, err := s.Foo(ctx, req.I, req.S)
		return FooResponse{I: i, Err: err}, nil
	}
}

type Endpoints struct {
	Foo endpoint.Endpoint
}
