package circuitbreaker

import (
	"context"

	"github.com/afex/hystrix-go/hystrix"

	"github.com/openmesh/kit/endpoint"
)

// Hystrix returns an endpoint.Middleware that implements the circuit
// breaker pattern using the afex/hystrix-go package.
//
// When using this circuit breaker, please configure your commands separately.
//
// See https://godoc.org/github.com/afex/hystrix-go/hystrix for more
// information.
func Hystrix[Request, Response any](commandName string) endpoint.Middleware[Request, Response] {
	return func(next endpoint.Endpoint[Request, Response]) endpoint.Endpoint[Request, Response] {
		return func(ctx context.Context, request Request) (response Response, err error) {
			var resp Response
			if err := hystrix.Do(commandName, func() (err error) {
				resp, err = next(ctx, request)
				return err
			}, nil); err != nil {
				return *new(Response), err
			}
			return resp, nil
		}
	}
}
