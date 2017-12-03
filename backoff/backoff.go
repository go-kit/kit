package backoff

import (
	"math/rand"
	"sync/atomic"
	"time"
)

const (
	DefaultInterval    = time.Second
	DefaultMaxInterval = time.Minute
)

// ExponentialBackoff provides jittered exponential durations for the purpose of
// avoiding flodding a service with requests.
type ExponentialBackoff struct {
	Interval time.Duration
	Max      time.Duration

	currentInterval atomic.Value
	cancel          <-chan struct{}
}

// New creates a new ExpontentialBackoff instance with the default values, and
// an optional cancel channel.
func New(cancel <-chan struct{}) *ExponentialBackoff {
	backoff := ExponentialBackoff{
		Interval: DefaultInterval,
		Max:      DefaultMaxInterval,
		cancel:   cancel,
	}
	backoff.Reset()
	return &backoff
}

// Reset should be called after a request succeeds.
func (b *ExponentialBackoff) Reset() {
	b.currentInterval.Store(b.Interval)
}

// Wait increases the backoff and blocks until the duration is over or the
// cancel channel is filled.
func (b *ExponentialBackoff) Wait() {
	d := b.NextBackoff()
	select {
	case <-time.After(d):
	case <-b.cancel:
	}
}

// NextBackoff updates the time interval and returns the updated value.
func (b *ExponentialBackoff) NextBackoff() time.Duration {
	d := b.next()
	if d > b.Max {
		d = b.Max
	}

	b.currentInterval.Store(d)
	return d
}

// next provides the exponential jittered backoff value. See
// https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/
// for rationale.
func (b *ExponentialBackoff) next() time.Duration {
	current := b.currentInterval.Load().(time.Duration)
	d := float64(current * 2)
	jitter := rand.Float64() + 0.5
	return time.Duration(d * jitter)
}
