package netrpc

import (
	"net/rpc"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/examples/addsvc/reqrep"
)

// NewClient takes a net/rpc Client that should point to an instance of an
// addsvc. It returns an endpoint that wraps and invokes that Client.
func NewClient(c *rpc.Client) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		var (
			errs      = make(chan error, 1)
			responses = make(chan interface{}, 1)
		)
		go func() {
			var response reqrep.AddResponse
			if err := c.Call("addsvc.Add", request, &response); err != nil {
				errs <- err
				return
			}
			responses <- response
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
