package ratelimit

import (
	"errors"
	"time"

	"github.com/tsenart/tb"
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
)

// ErrThrottled is returned in the request path when the rate limiter is
// triggered and the request is rejected.
var ErrThrottled = errors.New("throttled")

// NewTokenBucketThrottler returns an endpoint.Middleware that acts as a rate
// limiter based on a "token-bucket" algorithm. Requests that would exceed the
// maximum request rate are rejected with an error.
func NewTokenBucketThrottler(options ...TokenBucketOption) endpoint.Middleware {
	t := tokenBucketThrottler{
		freq: 100 * time.Millisecond,
		key:  "",
		rate: 100,
		take: 1,
	}
	for _, option := range options {
		option(&t)
	}
	throttler := tb.NewThrottler(t.freq)
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			if throttler.Halt(t.key, t.take, t.rate) {
				return nil, ErrThrottled
			}
			return next(ctx, request)
		}
	}
}

type tokenBucketThrottler struct {
	freq time.Duration
	key  string
	rate int64
	take int64
}

// TokenBucketOption sets an option on the token bucket throttler.
type TokenBucketOption func(*tokenBucketThrottler)

// TokenBucketFillFrequency sets the interval at which tokens are replenished
// into the bucket. By default, it's 100 milliseconds.
func TokenBucketFillFrequency(freq time.Duration) TokenBucketOption {
	return func(t *tokenBucketThrottler) { t.freq = freq }
}

// TokenBucketMaxRate sets the maximum allowed request rate.
// By default, it's 100.
func TokenBucketMaxRate(rate int64) TokenBucketOption {
	return func(t *tokenBucketThrottler) { t.rate = rate }
}

// TokenBucketTake sets the number of tokens taken with each request.
// By default, it's 1.
func TokenBucketTake(take int64) TokenBucketOption {
	return func(t *tokenBucketThrottler) { t.take = take }
}
