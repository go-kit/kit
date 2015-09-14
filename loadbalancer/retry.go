package loadbalancer

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
)

// Retry wraps the load balancer to make it behave like a simple endpoint.
// Requests to the endpoint will be automatically load balanced via the load
// balancer. Requests that return errors will be retried until they succeed,
// up to max times, or until the timeout is elapsed, whichever comes first.
func Retry(max int, timeout time.Duration, lb LoadBalancer) endpoint.Endpoint {
	if lb == nil {
		panic("nil LoadBalancer")
	}

	return func(ctx context.Context, request interface{}) (interface{}, error) {
		var (
			newctx, cancel = context.WithTimeout(ctx, timeout)
			responses      = make(chan interface{}, 1)
			errs           = make(chan error, 1)
			a              = []string{}
		)
		defer cancel()
		for i := 1; i <= max; i++ {
			go func() {
				e, err := lb.Endpoint()
				if err != nil {
					errs <- err
					return
				}
				response, err := e(newctx, request)
				if err != nil {
					errs <- err
					return
				}
				responses <- response
			}()

			select {
			case <-newctx.Done():
				return nil, newctx.Err()
			case response := <-responses:
				return response, nil
			case err := <-errs:
				a = append(a, err.Error())
				continue
			}
		}
		return nil, fmt.Errorf("retry attempts exceeded (%s)", strings.Join(a, "; "))
	}
}
