package zipkin

// Collector represents a Zipkin trace collector, which is probably a set of
// remote endpoints.
type Collector interface {
	Collect(*Span) error
}

// NopCollector implements Collector but performs no work.
type NopCollector struct{}

// Collect implements Collector.
func (NopCollector) Collect(*Span) error { return nil }
