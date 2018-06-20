package metrics

// Provider abstracts over metrics backends, and constructs metrics.
type Provider interface {
	NewCounter(Identifier) (Counter, error)
	NewGauge(Identifier) (Gauge, error)
	NewHistogram(Identifier) (Histogram, error)
}

// Identifier uniquely identifies a metric.
// Different backends may use different fields from the identifier.
type Identifier struct {
	// NameTemplate is used by the expvar and statsd providers. It supports
	// basic template interpolation. Strings surrounded by {} will be
	// interpreted as label keys and replaced with corresponding label values.
	//
	// For example, a NameTemplate `http_request_{method}_{code}_count` with
	// labels `{foo: bar, code: 200}` will be rendered as
	// `http_request_unknown_200_count`.
	NameTemplate string

	// Namespace is used by the Prometheus provider.
	Namespace string

	// Subsystem is used by the Prometheus provider.
	Subsystem string

	// Name is used by the Prometheus provider.
	Name string

	// Help is used by the Prometheus provider.
	Help string

	// Buckets is used by the Prometheus provider for histograms only.
	Buckets []float64

	// Labels are used by the Prometheus provider. All labels must be
	// predeclared when metrics are constructed.
	Labels []string
}
