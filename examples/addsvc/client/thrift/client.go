package thrift

import (
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	thriftadd "github.com/go-kit/kit/examples/addsvc/_thrift/gen-go/add"
	"github.com/go-kit/kit/examples/addsvc/reqrep"
)

// NewClient takes a Thrift AddServiceClient, which should point to an
// instance of an addsvc. It returns an endpoint that wraps and invokes that
// client.
func NewClient(client *thriftadd.AddServiceClient) endpoint.Endpoint {
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
			reply, err := client.Add(addReq.A, addReq.B)
			if err != nil {
				errs <- err
				return
			}
			responses <- reqrep.AddResponse{V: reply.Value}
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
