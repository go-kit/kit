package main

import (
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/examples/addsvc/pb"
	"github.com/go-kit/kit/examples/addsvc/reqrep"
)

// A binding wraps an Endpoint so that it's usable by a transport. grpcBinding
// makes an Endpoint usable over gRPC.
type grpcBinding struct{ endpoint.Endpoint }

// Add implements the proto3 AddServer by forwarding to the wrapped Endpoint.
//
// As far as I can tell, gRPC doesn't (currently) provide a user-accessible
// way to manipulate the RPC context, like headers for HTTP. So we don't have
// a way to transport e.g. Zipkin IDs with the request. TODO.
func (b grpcBinding) Add(ctx0 context.Context, req *pb.AddRequest) (*pb.AddReply, error) {
	var (
		ctx, cancel = context.WithCancel(ctx0)
		errs        = make(chan error, 1)
		replies     = make(chan *pb.AddReply, 1)
	)
	defer cancel()
	go func() {
		r, err := b.Endpoint(ctx, reqrep.AddRequest{A: req.A, B: req.B})
		if err != nil {
			errs <- err
			return
		}
		resp, ok := r.(reqrep.AddResponse)
		if !ok {
			errs <- endpoint.ErrBadCast
			return
		}
		replies <- &pb.AddReply{V: resp.V}
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
