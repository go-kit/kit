package circuitbreaker_test

import (
	"errors"
	"testing"

	"github.com/streadway/handy/breaker"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
)

func TestHandyBreaker(t *testing.T) {
	var (
		thru       = 0
		myError    = error(nil)
		ratio      = 0.05
		primeWith  = breaker.DefaultMinObservations * 10
		shouldPass = func(failed int) bool { return (float64(failed) / float64(primeWith+failed)) <= ratio }
		extraTries = 10
	)

	var e endpoint.Endpoint
	e = func(context.Context, interface{}) (interface{}, error) { thru++; return struct{}{}, myError }
	e = circuitbreaker.HandyBreaker(ratio)(e)

	// Prime with some successes.
	for i := 0; i < primeWith; i++ {
		if _, err := e(context.Background(), struct{}{}); err != nil {
			t.Fatal(err)
		}
	}

	// Now we start throwing errors.
	myError = errors.New(":(")

	// The first few should get thru.
	var letThru int
	for i := 0; shouldPass(i); i++ { // off-by-one
		letThru++
		if _, err := e(context.Background(), struct{}{}); err != myError {
			t.Fatalf("want %v, have %v", myError, err)
		}
	}

	// But the rest should be blocked by an open circuit.
	for i := 1; i <= extraTries; i++ {
		if _, err := e(context.Background(), struct{}{}); err != circuitbreaker.ErrCircuitBreakerOpen {
			t.Errorf("with request #%d, want %v, have %v", primeWith+letThru+i, circuitbreaker.ErrCircuitBreakerOpen, err)
		}
	}

	// Confirm the rest didn't get through.
	if want, have := primeWith+letThru, thru; want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}
