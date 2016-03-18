package zipkin

import "math"

// Sampler functions return if a Zipkin span should be sampled, based on its
// traceID.
type Sampler func(id int64) bool

// SampleRate returns a sampler function using a particular sample rate and a
// sample salt to identify if a Zipkin span based on its spanID should be
// collected.
func SampleRate(rate float64, salt int64) Sampler {
	if rate <= 0 {
		return func(_ int64) bool {
			return false
		}
	}
	if rate >= 1.0 {
		return func(_ int64) bool {
			return true
		}
	}
	return func(id int64) bool {
		return int64(math.Abs(float64(id^salt)))%10000 < int64(rate*10000)
	}
}
