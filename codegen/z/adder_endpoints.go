// Do not edit! Generated by gokit-generate

package z

import (
	"golang.org/x/net/context"
	"github.com/go-kit/kit/endpoint"
)

func MakeAdderEndpoints(x Adder) map[string]endpoint.Endpoint{
	m :=  map[string]endpoint.Endpoint{}

	m["Add"] = func (ctx context.Context, request interface{}) (interface{}, error) {
		select {
		default:
		case <-ctx.Done():
			return nil, endpoint.ErrContextCanceled
		}
		req, ok := request.(AdderAddRequest)
		if !ok {
			return nil, endpoint.ErrBadCast
		}
		var err error
		var resp AdderAddResponse
		resp.Int = x.Add(req.A, req.B)
		return resp, err
	}
	return m

}
