// Package ratelimit implements middleware for limiting
// the rate at which requests are executed.
package ratelimit

import (
	"errors"
	"time"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
)

// Bucket represents a token bucket.  See github.com/juju/ratelimit
// for one implementation.
type Bucket interface {
	// TakeMaxDuration takes count tokens from the bucket, but it
	// will only take tokens from the bucket if the wait time is no
	// greater than maxWait.
	//
	// If it would take longer than maxWait for the tokens to become
	// available, it does nothing and reports false, otherwise it
	// returns the time that the caller should wait until the tokens
	// are actually available, and reports true.
	TakeMaxDuration(count int64, maxWait time.Duration) (time.Duration, bool)
}

// timeSleep allows the tests to mock the implementation of time.Sleep.
var timeSleep = time.Sleep

// ErrLimited is returned in the request path when the rate limiter is
// triggered and the request is rejected.
var ErrLimited = errors.New("rate limit exceeded")

// NewTokenBucketLimiter returns an endpoint.Middleware that acts as a
// rate limiter using the given token bucket implementation. Each
// request takes one token from the bucket and waits for a maximum of
// maxSleep for a token to become available. If a token will not be
// available within that time, it will return an ErrLimited error.
//
// To make the request always return an error if the rate has been
// exceeded, pass 0 for maxSleep.
func NewTokenBucketLimiter(bucket Bucket, maxSleep time.Duration) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			d, ok := bucket.TakeMaxDuration(1, maxSleep)
			if !ok {
				return nil, ErrLimited
			}
			timeSleep(d)
			return next(ctx, request)
		}
	}
}
