package expvar

import "github.com/go-kit/kit/metrics2"

var (
	_ metrics.Provider  = (*Provider)(nil)
	_ metrics.Counter   = (*counter)(nil)
	_ metrics.Gauge     = (*gauge)(nil)
	_ metrics.Histogram = (*histogram)(nil)
)
