package main

import (
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	thriftadd "github.com/go-kit/kit/examples/addsvc/_thrift/gen-go/add"
	"github.com/go-kit/kit/examples/addsvc/reqrep"
)

// A binding wraps an Endpoint so that it's usable by a transport.
// thriftBinding makes an Endpoint usable over Thrift.
type thriftBinding struct {
	context.Context
	endpoint.Endpoint
}

// Add implements Thrift's AddService interface.
func (tb thriftBinding) Add(a, b int64) (*thriftadd.AddReply, error) {
	var (
		ctx, cancel = context.WithCancel(tb.Context)
		errs        = make(chan error, 1)
		replies     = make(chan *thriftadd.AddReply, 1)
	)
	defer cancel()
	go func() {
		r, err := tb.Endpoint(ctx, reqrep.AddRequest{A: a, B: b})
		if err != nil {
			errs <- err
			return
		}
		resp, ok := r.(reqrep.AddResponse)
		if !ok {
			errs <- endpoint.ErrBadCast
			return
		}
		replies <- &thriftadd.AddReply{Value: resp.V}
	}()
	select {
	case <-ctx.Done():
		return nil, context.DeadlineExceeded
	case err := <-errs:
		return nil, err
	case reply := <-replies:
		return reply, nil
	}
}
