package rpc

import (
	"net/rpc"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
)

// NewRPCClient takes a net/rpc Client that should point to an instance...
func NewRPCClient(c *rpc.Client) func(method string) endpoint.Endpoint {
	return func(method string) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			var (
				errs      = make(chan error, 1)
				responses = make(chan interface{}, 1)
			)
			go func() {
				var response ResponseT
				if err := c.Call(method, request, &response); err != nil {
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
}
