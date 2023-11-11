package circuitbreaker_test

import (
	"testing"

	handybreaker "github.com/streadway/handy/breaker"

	"github.com/openmesh/kit/circuitbreaker"
)

func TestHandyBreaker(t *testing.T) {
	var (
		failureRatio     = 0.05
		breaker          = circuitbreaker.HandyBreaker[interface{}, interface{}](handybreaker.NewBreaker(failureRatio))
		primeWith        = handybreaker.DefaultMinObservations * 10
		shouldPass       = func(n int) bool { return (float64(n) / float64(primeWith+n)) <= failureRatio }
		openCircuitError = handybreaker.ErrCircuitOpen.Error()
	)
	testFailingEndpoint(t, breaker, primeWith, shouldPass, 0, openCircuitError)
}
