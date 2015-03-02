package metrics

type scaledHistogram struct {
	Histogram
	scale int64
}

// NewScaledHistogram returns a Histogram whose observed values are downscaled
// (divided) by scale.
func NewScaledHistogram(h Histogram, scale int64) Histogram {
	return scaledHistogram{h, scale}
}

func (h scaledHistogram) With(f Field) Histogram {
	return scaledHistogram{
		Histogram: h.Histogram.With(f),
		scale:     h.scale,
	}
}

func (h scaledHistogram) Observe(value int64) {
	h.Histogram.Observe(value / h.scale)
}
