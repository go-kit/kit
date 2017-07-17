package histogram

import (
	"gopkg.in/caio/go-tdigest.v1"
)

type Histogram struct {
	d *tdigest.TDigest
}

func New() *Histogram {
	return &Histogram{
		d: tdigest.New(100),
	}
}

func (h *Histogram) Observe(value float64) {
	h.d.Add(value, 1)
}

func (h *Histogram) Quantile(q float64) float64 {
	return h.d.Quantile(q)
}
