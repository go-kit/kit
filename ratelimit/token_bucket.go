package ratelimit

import (
	"context"
	"errors"
	"time"

	"github.com/juju/ratelimit"

	"github.com/go-kit/kit/endpoint"
)

// ErrLimited is returned in the request path when the rate limiter is
// triggered and the request is rejected.
var ErrLimited = errors.New("rate limit exceeded")

// NewTokenBucketLimiter returns an endpoint.Middleware that acts as a rate
// limiter based on a token-bucket algorithm. Requests that would exceed the
// maximum request rate are simply rejected with an error.
func NewTokenBucketLimiter(tb *ratelimit.Bucket) endpoint.Middleware {
	return NewPerRequestTokenBucketLimiter(func(_ context.Context, _ interface{}) (*ratelimit.Bucket, error) {
		return tb, nil
	})
}

// NewPerRequestTokenBucketLimiter returns an endpoint.Middleware that acts as a
// rate limiter based on a token-bucket algorithm retrieved in runtime by a
// custom resolver defined by user. Requests that would exceed the maximum
// request rate are simply rejected with an error.
func NewPerRequestTokenBucketLimiter(resolver func(context.Context, interface{}) (*ratelimit.Bucket, error)) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			tb, err := resolver(ctx, request)
			if err != nil {
				return nil, err
			}
			if tb.TakeAvailable(1) == 0 {
				return nil, ErrLimited
			}
			return next(ctx, request)
		}
	}
}

// NewTokenBucketThrottler returns an endpoint.Middleware that acts as a
// request throttler based on a token-bucket algorithm. Requests that would
// exceed the maximum request rate are delayed via the parameterized sleep
// function. By default you may pass time.Sleep.
func NewTokenBucketThrottler(tb *ratelimit.Bucket, sleep func(time.Duration)) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			sleep(tb.Take(1))
			return next(ctx, request)
		}
	}
}
