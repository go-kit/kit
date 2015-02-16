package metrics

import "time"

// TimeHistogram is a convenience wrapper for a Histogram of time.Durations.
type TimeHistogram interface {
	With(Field) TimeHistogram
	Observe(time.Duration)
}

type timeHistogram struct {
	Histogram
	unit time.Duration
}

// NewTimeHistogram returns a TimeHistogram wrapper around the passed
// Histogram, in units of unit.
func NewTimeHistogram(h Histogram, unit time.Duration) TimeHistogram {
	return &timeHistogram{
		Histogram: h,
		unit:      unit,
	}
}

func (h *timeHistogram) With(f Field) TimeHistogram {
	return &timeHistogram{
		Histogram: h.Histogram.With(f),
		unit:      h.unit,
	}
}

func (h *timeHistogram) Observe(d time.Duration) {
	h.Histogram.Observe(int64(d / h.unit))
}
