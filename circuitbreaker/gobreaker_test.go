package circuitbreaker_test

import (
	"errors"
	"testing"
	"time"

	"github.com/sony/gobreaker"
	"golang.org/x/net/context"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
)

func TestGobreaker(t *testing.T) {
	var (
		thru        int
		last        gobreaker.State
		myError     = errors.New("❤️")
		timeout     = time.Millisecond
		stateChange = func(_ string, from, to gobreaker.State) { last = to }
	)

	var e endpoint.Endpoint
	e = func(context.Context, interface{}) (interface{}, error) { thru++; return struct{}{}, myError }
	e = circuitbreaker.Gobreaker(gobreaker.Settings{
		Timeout:       timeout,
		OnStateChange: stateChange,
	})(e)

	// "Default ReadyToTrip returns true when the number of consecutive
	// failures is more than 5."
	// https://github.com/sony/gobreaker/blob/bfa846d/gobreaker.go#L76
	for i := 0; i < 5; i++ {
		if _, err := e(context.Background(), struct{}{}); err != myError {
			t.Errorf("want %v, have %v", myError, err)
		}
	}

	if want, have := 5, thru; want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	e(context.Background(), struct{}{})
	if want, have := 6, thru; want != have { // got thru
		t.Errorf("want %d, have %d", want, have)
	}
	if want, have := gobreaker.StateOpen, last; want != have { // tripped
		t.Errorf("want %v, have %v", want, have)
	}

	e(context.Background(), struct{}{})
	if want, have := 6, thru; want != have { // didn't get thru
		t.Errorf("want %d, have %d", want, have)
	}

	time.Sleep(2 * timeout)

	e(context.Background(), struct{}{})
	if want, have := 7, thru; want != have { // got thru via halfopen
		t.Errorf("want %d, have %d", want, have)
	}
	if want, have := gobreaker.StateOpen, last; want != have { // re-tripped
		t.Errorf("want %v, have %v", want, have)
	}

	time.Sleep(2 * timeout)

	myError = nil
	e(context.Background(), struct{}{})
	if want, have := 8, thru; want != have { // got thru via halfopen
		t.Errorf("want %d, have %d", want, have)
	}
	if want, have := gobreaker.StateClosed, last; want != have { // now it's good
		t.Errorf("want %v, have %v", want, have)
	}
}
