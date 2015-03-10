package metrics

type multiCounter []Counter

// NewMultiCounter returns a wrapper around multiple Counters.
func NewMultiCounter(counters ...Counter) Counter {
	c := make(multiCounter, 0, len(counters))
	for _, counter := range counters {
		c = append(c, counter)
	}
	return c
}

func (c multiCounter) With(f Field) Counter {
	next := make(multiCounter, len(c))
	for i, counter := range c {
		next[i] = counter.With(f)
	}
	return next
}

func (c multiCounter) Add(delta uint64) {
	for _, counter := range c {
		counter.Add(delta)
	}
}

type multiGauge []Gauge

// NewMultiGauge returns a wrapper around multiple Gauges.
func NewMultiGauge(gauges ...Gauge) Gauge {
	g := make(multiGauge, 0, len(gauges))
	for _, gauge := range gauges {
		g = append(g, gauge)
	}
	return g
}

func (g multiGauge) With(f Field) Gauge {
	next := make(multiGauge, len(g))
	for i, gauge := range g {
		next[i] = gauge.With(f)
	}
	return next
}

func (g multiGauge) Set(value float64) {
	for _, gauge := range g {
		gauge.Set(value)
	}
}

func (g multiGauge) Add(delta float64) {
	for _, gauge := range g {
		gauge.Add(delta)
	}
}

type multiHistogram []Histogram

// NewMultiHistogram returns a wrapper around multiple Histograms.
func NewMultiHistogram(histograms ...Histogram) Histogram {
	h := make(multiHistogram, 0, len(histograms))
	for _, histogram := range histograms {
		h = append(h, histogram)
	}
	return h
}

func (h multiHistogram) With(f Field) Histogram {
	next := make(multiHistogram, len(h))
	for i, histogram := range h {
		next[i] = histogram.With(f)
	}
	return next
}

func (h multiHistogram) Observe(value int64) {
	for _, histogram := range h {
		histogram.Observe(value)
	}
}
