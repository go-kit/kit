package main

import (
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/examples/addsvc/reqrep"
)

// NetrpcBinding makes an endpoint usable over net/rpc. It needs to be
// exported to be picked up by net/rpc.
type NetrpcBinding struct {
	ctx context.Context // has methods which should not be made available
	endpoint.Endpoint
}

// Add implements the net/rpc method definition.
func (b NetrpcBinding) Add(request reqrep.AddRequest, response *reqrep.AddResponse) error {
	var (
		ctx, cancel = context.WithCancel(b.ctx)
		errs        = make(chan error, 1)
		responses   = make(chan reqrep.AddResponse, 1)
	)
	defer cancel()
	go func() {
		resp, err := b.Endpoint(ctx, request)
		if err != nil {
			errs <- err
			return
		}
		addResp, ok := resp.(reqrep.AddResponse)
		if !ok {
			errs <- endpoint.ErrBadCast
			return
		}
		responses <- addResp
	}()
	select {
	case <-ctx.Done():
		return context.DeadlineExceeded
	case err := <-errs:
		return err
	case resp := <-responses:
		(*response) = resp
		return nil
	}
}
