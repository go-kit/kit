package circuitbreaker

import (
	"github.com/sony/gobreaker"
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
)

// NewSonyCircuitBreaker returns an endpoint.Middleware that permits the
// request if the underlying circuit breaker allows it. Only errors returned
// by the wrapped endpoint count against the circuit breaker's error count.
// See github.com/sony/gobreaker for more information.
func NewSonyCircuitBreaker(settings gobreaker.Settings) endpoint.Middleware {
	cb := gobreaker.NewCircuitBreaker(settings)
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			return cb.Execute(func() (interface{}, error) { return next(ctx, request) })
		}
	}
}
