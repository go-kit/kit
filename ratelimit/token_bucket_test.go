package ratelimit_test

import (
	"context"
	"strings"
	"testing"
	"time"

	jujuratelimit "github.com/juju/ratelimit"
	"golang.org/x/time/rate"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/ratelimit"
)

var nopEndpoint = func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil }

func TestTokenBucketLimiter(t *testing.T) {
	tb := jujuratelimit.NewBucket(time.Minute, 1)
	testSuccessThenFailure(
		t,
		ratelimit.NewTokenBucketLimiter(tb)(nopEndpoint),
		ratelimit.ErrLimited.Error())
}

func TestTokenBucketThrottler(t *testing.T) {
	tb := jujuratelimit.NewBucket(time.Minute, 1)
	testSuccessThenFailure(
		t,
		ratelimit.NewTokenBucketThrottler(tb, nil)(nopEndpoint),
		"context deadline exceeded")
}

func TestXRateErroring(t *testing.T) {
	limit := rate.NewLimiter(rate.Every(time.Minute), 1)
	testSuccessThenFailure(
		t,
		ratelimit.NewErroringLimiter(limit)(nopEndpoint),
		ratelimit.ErrLimited.Error())
}

func TestXRateDelaying(t *testing.T) {
	limit := rate.NewLimiter(rate.Every(time.Minute), 1)
	testSuccessThenFailure(
		t,
		ratelimit.NewDelayingLimiter(limit)(nopEndpoint),
		"exceed context deadline")
}

func testSuccessThenFailure(t *testing.T, e endpoint.Endpoint, failContains string) {
	ctx, cxl := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cxl()

	// First request should succeed.
	if _, err := e(ctx, struct{}{}); err != nil {
		t.Errorf("unexpected: %v\n", err)
	}

	// Next request should fail.
	if _, err := e(ctx, struct{}{}); !strings.Contains(err.Error(), failContains) {
		t.Errorf("expected `%s`: %v\n", failContains, err)
	}
}
