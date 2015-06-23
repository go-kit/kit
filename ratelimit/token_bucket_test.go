package ratelimit_test

import (
	"math"
	"testing"
	"time"

	jujuratelimit "github.com/juju/ratelimit"
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/ratelimit"
)

func TestTokenBucketLimiter(t *testing.T) {
	for _, n := range []int{1, 2, 100} {
		b := jujuratelimit.NewBucketWithRate(float64(n), int64(n))
		limiter := ratelimit.NewTokenBucketLimiter(b, 0)
		testLimiter(t, limiter(noNext), n)
	}
}

func testLimiter(t *testing.T, e endpoint.Endpoint, rate int) {
	// First <rate> requests should succeed.
	for i := 0; i < rate; i++ {
		if _, err := e(context.Background(), struct{}{}); err != nil {
			t.Fatalf("rate=%d: request %d/%d failed: %v", rate, i+1, rate, err)
		}
	}

	// Next request should fail.
	if _, err := e(context.Background(), struct{}{}); err != ratelimit.ErrLimited {
		t.Errorf("rate=%d: want %v, have %v", rate, ratelimit.ErrLimited, err)
	}
}

func TestTokenBucketThrottling(t *testing.T) {
	var d time.Duration
	*ratelimit.TimeSleep = func(d0 time.Duration) { d = d0 }
	defer func() {
		*ratelimit.TimeSleep = time.Sleep
	}()
	b := jujuratelimit.NewBucketWithRate(1, 1)
	e := ratelimit.NewTokenBucketLimiter(b, 10*time.Second)(noNext)

	// First request should go through with no delay.
	_, err := e(context.Background(), struct{}{})
	if err != nil {
		t.Fatalf("unexpected error from request: %v", err)
	}
	if want, have := time.Duration(0), d; want != have {
		t.Errorf("want %s, have %s", want, have)
	}

	// Next request should request a ~1s sleep.
	_, err = e(context.Background(), struct{}{})
	if err != nil {
		t.Fatalf("unexpected error from request: %v", err)
	}
	if want, have, tol := time.Second, d, time.Millisecond; math.Abs(float64(want-have)) > float64(tol) {
		t.Errorf("want %s, have %s", want, have)
	}
}

func TestTokenBucketThrottlerTimeout(t *testing.T) {
	var d time.Duration
	*ratelimit.TimeSleep = func(d0 time.Duration) { d = d0 }
	defer func() {
		*ratelimit.TimeSleep = time.Sleep
	}()
	b := jujuratelimit.NewBucketWithRate(1, 1)
	e := ratelimit.NewTokenBucketLimiter(b, 500*time.Millisecond)(noNext)

	// First request should go through with no delay.
	_, err := e(context.Background(), struct{}{})
	if err != nil {
		t.Fatalf("unexpected error from request: %v", err)
	}
	if want, have := time.Duration(0), d; want != have {
		t.Errorf("want %s, have %s", want, have)
	}

	// Next request should fail because it would need to
	// wait for >500ms.
	_, err = e(context.Background(), struct{}{})
	if _, err := e(context.Background(), struct{}{}); err != ratelimit.ErrLimited {
		t.Errorf("want %v, have %v", ratelimit.ErrLimited, err)
	}
}

func noNext(context.Context, interface{}) (interface{}, error) {
	return struct{}{}, nil
}
