package histogram

type Histogram struct{}

func New() *Histogram {
	return &Histogram{}
}

func (*Histogram) Observe(value float64) {}

func (*Histogram) Quantile(q float64) float64 {
	return 0.0
}
