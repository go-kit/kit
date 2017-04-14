package metrics

// Provider abstracts over constructors and lifecycle management functions for
// each supported metrics backend. It should only be used by those who need to
// swap out implementations dynamically.
//
// This is primarily useful for intermediating frameworks, and is likely
// unnecessary for most Go kit services. See the provider package-level doc
// comment for more typical usage instructions.
type Provider interface {
	NewCounter(name string) Counter
	NewGauge(name string) Gauge
	NewHistogram(name string, buckets int) Histogram
	Stop()
}
