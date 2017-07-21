package ratelimit_test

import (
	"context"
	"math"
	"strings"
	"testing"
	"time"

	jujuratelimit "github.com/juju/ratelimit"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/ratelimit"
	"golang.org/x/time/rate"
)

func TestTokenBucketLimiter(t *testing.T) {
	e := func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil }
	for _, n := range []int{1, 2, 100} {
		tb := jujuratelimit.NewBucketWithRate(float64(n), int64(n))
		testLimiter(t, ratelimit.NewTokenBucketLimiter(tb)(e), n)
	}
}

func TestTokenBucketThrottler(t *testing.T) {
	d := time.Duration(0)
	s := func(d0 time.Duration) { d = d0 }

	e := func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil }
	e = ratelimit.NewTokenBucketThrottler(jujuratelimit.NewBucketWithRate(1, 1), s)(e)

	// First request should go through with no delay.
	e(context.Background(), struct{}{})
	if want, have := time.Duration(0), d; want != have {
		t.Errorf("want %s, have %s", want, have)
	}

	// Next request should request a ~1s sleep.
	e(context.Background(), struct{}{})
	if want, have, tol := time.Second, d, time.Millisecond; math.Abs(float64(want-have)) > float64(tol) {
		t.Errorf("want %s, have %s", want, have)
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

func TestXRateErroring(t *testing.T) {
	limit := rate.NewLimiter(rate.Every(time.Minute), 1)
	e := func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil }
	testLimiter(t, ratelimit.NewErroringLimiter(limit)(e), 1)
}

func TestXRateDelaying(t *testing.T) {
	limit := rate.NewLimiter(rate.Every(time.Minute), 1)
	e := func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil }
	e = ratelimit.NewDelayingLimiter(limit)(e)

	_, err := e(context.Background(), struct{}{})
	if err != nil {
		t.Errorf("unexpected: %v\n", err)
	}

	dur := 500 * time.Millisecond
	ctx, cxl := context.WithTimeout(context.Background(), dur)
	defer cxl()

	_, err = e(ctx, struct{}{})
	if !strings.Contains(err.Error(), "exceed context deadline") {
		t.Errorf("expected timeout: %v\n", err)
	}
}
