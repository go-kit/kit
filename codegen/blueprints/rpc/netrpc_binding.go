package rpc

import (
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
)

// NetrpcBinding makes an endpoint usable over net/rpc. It needs to be
// exported to be picked up by net/rpc.
type NetrpcBinding struct {
	Ctx context.Context // has methods which should not be made available
	endpoint.Endpoint
}

type RequestT struct{}
type ResponseT struct{}

// Fun implements the net/rpc method definition.
func (b NetrpcBinding) FunT(request RequestT, response *ResponseT) error {
	var (
		ctx, cancel = context.WithCancel(b.Ctx)
		errs        = make(chan error, 1)
		responses   = make(chan ResponseT, 1)
	)
	defer cancel()
	go func() {
		rawResp, err := b.Endpoint(ctx, request)
		if err != nil {
			errs <- err
			return
		}
		resp, ok := rawResp.(ResponseT)
		if !ok {
			errs <- endpoint.ErrBadCast
			return
		}
		responses <- resp
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
