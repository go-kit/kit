package expvar

import "github.com/go-kit/kit/metrics2"

var (
	_ metrics.Provider  = (*Provider)(nil)
	_ metrics.Counter   = (*IntCounter)(nil)
	_ metrics.Counter   = (*FloatCounter)(nil)
	_ metrics.Gauge     = (*IntGauge)(nil)
	_ metrics.Gauge     = (*FloatGauge)(nil)
	_ metrics.Histogram = (*Histogram)(nil)
)
