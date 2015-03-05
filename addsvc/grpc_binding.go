package main

import (
	"github.com/peterbourgon/gokit/addsvc/pb"
	"github.com/peterbourgon/gokit/server"
	"golang.org/x/net/context"
)

type grpcBinding struct{ server.Endpoint }

// Add implements the proto3 AddServer by forwarding to the wrapped Endpoint.
func (b grpcBinding) Add(ctx context.Context, req *pb.AddRequest) (*pb.AddReply, error) {
	addReq := request{req.A, req.B}
	r, err := b.Endpoint(ctx, addReq)
	if err != nil {
		return nil, err
	}

	resp, ok := r.(*response)
	if !ok {
		return nil, server.ErrBadCast
	}

	return &pb.AddReply{
		V: resp.V,
	}, nil
}
