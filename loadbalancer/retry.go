package loadbalancer

import (
	"fmt"
	"strings"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
)

// Retry yields an endpoint that invokes the load balancer up to max times.
func Retry(max int, lb LoadBalancer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		errs := []string{}
		for i := 1; i <= max; i++ {
			e, err := lb.Get()
			if err != nil {
				errs = append(errs, err.Error())
				return nil, fmt.Errorf("%s", strings.Join(errs, "; ")) // fatal
			}
			response, err := e(ctx, request)
			if err != nil {
				errs = append(errs, err.Error())
				continue // try again
			}
			return response, err
		}
		if len(errs) <= 0 {
			panic("impossible state in retry load balancer")
		}
		return nil, fmt.Errorf("%s", strings.Join(errs, "; "))
	}
}
