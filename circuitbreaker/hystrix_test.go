package circuitbreaker_test

import (
	"errors"
	"testing"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"golang.org/x/net/context"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
)

func TestHystrixCircuitBreakerOpen(t *testing.T) {
	var (
		thru       = 0
		myError    = error(nil)
		ratio      = 0.04
		primeWith  = hystrix.DefaultVolumeThreshold * 2
		shouldPass = func(failed int) bool { return (float64(failed) / float64(primeWith+failed)) <= ratio }
		extraTries = 10
	)

	// configure hystrix
	hystrix.ConfigureCommand("myEndpoint", hystrix.CommandConfig{
		ErrorPercentThreshold: 5,
		MaxConcurrentRequests: 200,
	})

	var e endpoint.Endpoint
	e = func(context.Context, interface{}) (interface{}, error) { thru++; return struct{}{}, myError }
	e = circuitbreaker.Hystrix("myEndpoint")(e)

	// prime
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
		if _, err := e(context.Background(), struct{}{}); err != hystrix.ErrCircuitOpen {
			t.Errorf("with request #%d, want %v, have %v", primeWith+letThru+i, hystrix.ErrCircuitOpen, err)
		}
	}

	// Confirm the rest didn't get through.
	if want, have := primeWith+letThru, thru; want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestHystrixTimeout(t *testing.T) {
	var (
		timeout    = time.Millisecond * 0
		primeWith  = hystrix.DefaultVolumeThreshold * 2
		failNumber = 2 // 5% threshold
	)

	// configure hystrix
	hystrix.ConfigureCommand("timeoutEndpoint", hystrix.CommandConfig{
		ErrorPercentThreshold: 5,
		MaxConcurrentRequests: 200,
		SleepWindow:           5, // milliseconds
		Timeout:               1, // milliseconds
	})

	var e endpoint.Endpoint
	e = func(context.Context, interface{}) (interface{}, error) {
		time.Sleep(2 * timeout)
		return struct{}{}, nil
	}
	e = circuitbreaker.Hystrix("timeoutEndpoint")(e)

	// prime
	for i := 0; i < primeWith; i++ {
		if _, err := e(context.Background(), struct{}{}); err != nil {
			t.Errorf("expecting %v, have %v", nil, err)
		}
	}

	// times out
	timeout = time.Millisecond * 2
	for i := 0; i < failNumber; i++ {
		if _, err := e(context.Background(), struct{}{}); err != hystrix.ErrTimeout {
			t.Errorf("%d expecting %v, have %v", i, hystrix.ErrTimeout, err)
		}
	}

	// fix timeout
	timeout = time.Millisecond * 0

	// fails for a little while still
	for i := 0; i < failNumber; i++ {
		if _, err := e(context.Background(), struct{}{}); err != hystrix.ErrCircuitOpen {
			t.Errorf("expecting %v, have %v", hystrix.ErrCircuitOpen, err)
		}
	}

	// back to OK
	time.Sleep(time.Millisecond * 5)
	if _, err := e(context.Background(), struct{}{}); err != nil {
		t.Errorf("expecting %v, have %v", nil, err)
	}
}
