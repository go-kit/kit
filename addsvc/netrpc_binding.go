package main

import (
	"golang.org/x/net/context"

	"github.com/go-kit/kit/addsvc/reqrep"
	"github.com/go-kit/kit/endpoint"
)

// A binding wraps an Endpoint so that it's usable by a transport. NetrpcBinding
// makes an Endpoint usable over net/rpc. It needs to be exported to be usable.
type NetrpcBinding struct {
	ctx context.Context // this has methods which should not be available
	endpoint.Endpoint
}

// Add implements the net/rpc method definition.
func (b NetrpcBinding) Add(request reqrep.AddRequest, response *reqrep.AddResponse) error {
	resp, err := b.Endpoint(b.ctx, request)
	if err != nil {
		return err
	}
	addResp, ok := resp.(reqrep.AddResponse)
	if !ok {
		return endpoint.ErrBadCast
	}
	(*response) = addResp
	return nil
}
