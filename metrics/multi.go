package metrics

type multiCounter struct {
	name string
	a    []Counter
}

// NewMultiCounter returns a wrapper around multiple Counters.
func NewMultiCounter(name string, counters ...Counter) Counter {
	return &multiCounter{
		name: name,
		a:    counters,
	}
}

func (c multiCounter) Name() string { return c.name }

func (c multiCounter) With(f Field) Counter {
	next := &multiCounter{
		name: c.name,
		a:    make([]Counter, len(c.a)),
	}
	for i, counter := range c.a {
		next.a[i] = counter.With(f)
	}
	return next
}

func (c multiCounter) Add(delta uint64) {
	for _, counter := range c.a {
		counter.Add(delta)
	}
}

type multiGauge struct {
	name string
	a    []Gauge
}

func (g multiGauge) Name() string { return g.name }

// NewMultiGauge returns a wrapper around multiple Gauges.
func NewMultiGauge(name string, gauges ...Gauge) Gauge {
	return &multiGauge{
		name: name,
		a:    gauges,
	}
}

func (g multiGauge) With(f Field) Gauge {
	next := &multiGauge{
		name: g.name,
		a:    make([]Gauge, len(g.a)),
	}
	for i, gauge := range g.a {
		next.a[i] = gauge.With(f)
	}
	return next
}

func (g multiGauge) Set(value float64) {
	for _, gauge := range g.a {
		gauge.Set(value)
	}
}

func (g multiGauge) Add(delta float64) {
	for _, gauge := range g.a {
		gauge.Add(delta)
	}
}

func (g multiGauge) Get() float64 {
	panic("cannot call Get on a MultiGauge")
}

type multiHistogram struct {
	name string
	a    []Histogram
}

// NewMultiHistogram returns a wrapper around multiple Histograms.
func NewMultiHistogram(name string, histograms ...Histogram) Histogram {
	return &multiHistogram{
		name: name,
		a:    histograms,
	}
}

func (h multiHistogram) Name() string { return h.name }

func (h multiHistogram) With(f Field) Histogram {
	next := &multiHistogram{
		name: h.name,
		a:    make([]Histogram, len(h.a)),
	}
	for i, histogram := range h.a {
		next.a[i] = histogram.With(f)
	}
	return next
}

func (h multiHistogram) Observe(value int64) {
	for _, histogram := range h.a {
		histogram.Observe(value)
	}
}

func (h multiHistogram) Distribution() ([]Bucket, []Quantile) {
	// TODO(pb): there may be a way to do this
	panic("cannot call Distribution on a MultiHistogram")
}
