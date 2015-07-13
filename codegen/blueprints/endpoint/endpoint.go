package endpoint

import (
	"github.com/go-kit/kit/endpoint"
	"golang.org/x/net/context"
)

type RequestT struct {
	A int
	B string
}

type ResponseT struct {
	P string
	Q float64
}

type FunT func(ctx context.Context, a int, b string) (p string, q float64, err error)

// makeEndpoint returns an endpoint wrapping the passed Add. If Add were an
// interface with multiple methods, we'd need individual endpoints for each.

// This function is just boiler-plate; in theory, it could be generated.

// map[string]endpoint
func makeEndpoint(f FunT) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		select {
		default:
		case <-ctx.Done():
			return nil, endpoint.ErrContextCanceled
		}

		req, ok := request.(RequestT)
		if !ok {
			return nil, endpoint.ErrBadCast
		}
		var r ResponseT
		var err error
		r.P, r.Q, err = f(ctx, req.A, req.B)
		return r, err
	}
}
