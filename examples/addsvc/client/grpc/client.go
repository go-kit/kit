package grpc

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/examples/addsvc/pb"
	"github.com/go-kit/kit/examples/addsvc/reqrep"
)

// NewClient takes a gRPC ClientConn that should point to an instance of an
// addsvc. It returns an endpoint that wraps and invokes that ClientConn.
func NewClient(cc *grpc.ClientConn) endpoint.Endpoint {
	client := pb.NewAddClient(cc)
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		var (
			errs      = make(chan error, 1)
			responses = make(chan interface{}, 1)
		)
		go func() {
			addReq, ok := request.(reqrep.AddRequest)
			if !ok {
				errs <- endpoint.ErrBadCast
				return
			}
			reply, err := client.Add(ctx, &pb.AddRequest{A: addReq.A, B: addReq.B})
			if err != nil {
				errs <- err
				return
			}
			responses <- reqrep.AddResponse{V: reply.V}
		}()
		select {
		case <-ctx.Done():
			return nil, context.DeadlineExceeded
		case err := <-errs:
			return nil, err
		case response := <-responses:
			return response, nil
		}
	}
}
