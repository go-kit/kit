package grpc

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/go-kit/kit/addsvc/pb"
	"github.com/go-kit/kit/addsvc/reqrep"
	"github.com/go-kit/kit/endpoint"
)

// NewClient takes a gRPC ClientConn that should point to an instance of an
// addsvc. It returns an endpoint that wraps and invokes that ClientConn.
func NewClient(cc *grpc.ClientConn) endpoint.Endpoint {
	client := pb.NewAddClient(cc)
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		addReq, ok := request.(reqrep.AddRequest)
		if !ok {
			return nil, endpoint.ErrBadCast
		}
		reply, err := client.Add(ctx, &pb.AddRequest{A: addReq.A, B: addReq.B})
		if err != nil {
			return nil, err
		}
		return reqrep.AddResponse{V: reply.V}, nil
	}
}
