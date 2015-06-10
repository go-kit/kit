package ratelimit_test

import (
	"testing"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/ratelimit"
)

func TestTokenBucketThrottler(t *testing.T) {
	e := func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil }
	testRateLimit(t, ratelimit.NewTokenBucketThrottler(ratelimit.TokenBucketMaxRate(0))(e), 0)     // all fail
	testRateLimit(t, ratelimit.NewTokenBucketThrottler(ratelimit.TokenBucketMaxRate(1))(e), 1)     // first pass
	testRateLimit(t, ratelimit.NewTokenBucketThrottler(ratelimit.TokenBucketMaxRate(100))(e), 100) // 100 pass
}

func testRateLimit(t *testing.T, e endpoint.Endpoint, rate int) {
	ctx := context.Background()
	for i := 0; i < rate; i++ {
		if _, err := e(ctx, struct{}{}); err != nil {
			t.Fatalf("rate=%d: request %d/%d failed: %v", rate, i+1, rate, err)
		}
	}
	if _, err := e(ctx, struct{}{}); err != ratelimit.ErrThrottled {
		t.Errorf("rate=%d: want %v, have %v", rate, ratelimit.ErrThrottled, err)
	}
}
