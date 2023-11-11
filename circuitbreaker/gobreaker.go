package circuitbreaker

import (
	"context"

	"github.com/sony/gobreaker"

	"github.com/openmesh/kit/endpoint"
)

// Gobreaker returns an endpoint.Middleware that implements the circuit
// breaker pattern using the sony/gobreaker package. Only errors returned by
// the wrapped endpoint count against the circuit breaker's error count.
//
// See http://godoc.org/github.com/sony/gobreaker for more information.
func Gobreaker[Request, Response any](cb *gobreaker.CircuitBreaker) endpoint.Middleware[Request, Response] {
	return func(next endpoint.Endpoint[Request, Response]) endpoint.Endpoint[Request, Response] {
		return func(ctx context.Context, request Request) (Response, error) {
			res, err := cb.Execute(func() (interface{}, error) { return next(ctx, request) })
			return res.(Response), err
		}
	}
}
