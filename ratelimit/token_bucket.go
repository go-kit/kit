package ratelimit

import (
	"errors"
	"time"

	juju "github.com/juju/ratelimit"
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
)

// ErrLimited is returned in the request path when the rate limiter is
// triggered and the request is rejected.
var ErrLimited = errors.New("rate limit exceeded")

// NewTokenBucketLimiter returns an endpoint.Middleware that acts as a rate
// limiter based on a token-bucket algorithm. Requests that would exceed the
// maximum request rate are simply rejected with an error.
func NewTokenBucketLimiter(options ...TokenBucketLimiterOption) endpoint.Middleware {
	limiter := tokenBucketLimiter{
		rate:     100,
		capacity: 100,
		take:     1,
	}
	for _, option := range options {
		option(&limiter)
	}
	tb := juju.NewBucketWithRate(limiter.rate, limiter.capacity)
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			if tb.TakeAvailable(limiter.take) == 0 {
				return nil, ErrLimited
			}
			return next(ctx, request)
		}
	}
}

type tokenBucketLimiter struct {
	rate     float64
	capacity int64
	take     int64
}

// TokenBucketLimiterOption sets a parameter on the TokenBucketLimiter.
type TokenBucketLimiterOption func(*tokenBucketLimiter)

// TokenBucketLimiterRate sets the rate (per second) at which tokens are
// replenished into the bucket. For most use cases, this should be the same as
// the capacity. By default, the rate is 100.
func TokenBucketLimiterRate(rate float64) TokenBucketLimiterOption {
	return func(tb *tokenBucketLimiter) { tb.rate = rate }
}

// TokenBucketLimiterCapacity sets the maximum number of tokens that the
// bucket will hold. For most use cases, this should be the same as the rate.
// By default, the capacity is 100.
func TokenBucketLimiterCapacity(capacity int64) TokenBucketLimiterOption {
	return func(tb *tokenBucketLimiter) { tb.capacity = capacity }
}

// TokenBucketLimiterTake sets the number of tokens that will be taken from
// the bucket with each request. By default, this is 1.
func TokenBucketLimiterTake(take int64) TokenBucketLimiterOption {
	return func(tb *tokenBucketLimiter) { tb.take = take }
}

// NewTokenBucketThrottler returns an endpoint.Middleware that acts as a
// request throttler based on a token-bucket algorithm. Requests that would
// exceed the maximum request rate are delayed via a parameterized sleep
// function.
func NewTokenBucketThrottler(options ...TokenBucketThrottlerOption) endpoint.Middleware {
	throttler := tokenBucketThrottler{
		tokenBucketLimiter: tokenBucketLimiter{
			rate:     100,
			capacity: 100,
			take:     1,
		},
		sleep: time.Sleep,
	}
	for _, option := range options {
		option(&throttler)
	}
	tb := juju.NewBucketWithRate(throttler.rate, throttler.capacity)
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			throttler.sleep(tb.Take(throttler.take))
			return next(ctx, request)
		}
	}
}

type tokenBucketThrottler struct {
	tokenBucketLimiter
	sleep func(time.Duration)
}

// TokenBucketThrottlerOption sets a parameter on the TokenBucketThrottler.
type TokenBucketThrottlerOption func(*tokenBucketThrottler)

// TokenBucketThrottlerRate sets the rate (per second) at which tokens are
// replenished into the bucket. For most use cases, this should be the same as
// the capacity. By default, the rate is 100.
func TokenBucketThrottlerRate(rate float64) TokenBucketThrottlerOption {
	return func(tb *tokenBucketThrottler) { tb.rate = rate }
}

// TokenBucketThrottlerCapacity sets the maximum number of tokens that the
// bucket will hold. For most use cases, this should be the same as the rate.
// By default, the capacity is 100.
func TokenBucketThrottlerCapacity(capacity int64) TokenBucketThrottlerOption {
	return func(tb *tokenBucketThrottler) { tb.capacity = capacity }
}

// TokenBucketThrottlerTake sets the number of tokens that will be taken from
// the bucket with each request. By default, this is 1.
func TokenBucketThrottlerTake(take int64) TokenBucketThrottlerOption {
	return func(tb *tokenBucketThrottler) { tb.take = take }
}

// TokenBucketThrottlerSleep sets the sleep function that's invoked to
// throttle requests. By default, it's time.Sleep.
func TokenBucketThrottlerSleep(sleep func(time.Duration)) TokenBucketThrottlerOption {
	return func(tb *tokenBucketThrottler) { tb.sleep = sleep }
}
