package circuitbreaker

import (
	"context"
	"time"

	"github.com/streadway/handy/breaker"

	"github.com/openmesh/kit/endpoint"
)

// HandyBreaker returns an endpoint.Middleware that implements the circuit
// breaker pattern using the streadway/handy/breaker package. Only errors
// returned by the wrapped endpoint count against the circuit breaker's error
// count.
//
// See http://godoc.org/github.com/streadway/handy/breaker for more
// information.
func HandyBreaker[Request, Response any](cb breaker.Breaker) endpoint.Middleware[Request, Response] {
	return func(next endpoint.Endpoint[Request, Response]) endpoint.Endpoint[Request, Response] {
		return func(ctx context.Context, request Request) (response Response, err error) {
			if !cb.Allow() {
				return *new(Response), breaker.ErrCircuitOpen
			}

			defer func(begin time.Time) {
				if err == nil {
					cb.Success(time.Since(begin))
				} else {
					cb.Failure(time.Since(begin))
				}
			}(time.Now())

			response, err = next(ctx, request)
			return
		}
	}
}
