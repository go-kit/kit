package expvar

import (
	"testing"

	"github.com/go-kit/kit/metrics2"
)

var (
	_ metrics.Counter   = (*IntCounter)(nil)
	_ metrics.Counter   = (*FloatCounter)(nil)
	_ metrics.Gauge     = (*IntGauge)(nil)
	_ metrics.Gauge     = (*FloatGauge)(nil)
	_ metrics.Histogram = (*Histogram)(nil)
)

func TestFieldIsolation(t *testing.T) {
	t.Parallel()

	p := NewProvider()
	c0 := p.NewIntCounter("foo")
	c1 := c0.With("key", "val").(*IntCounter)

	if want, have := 0, len(c0.fields); want != have {
		t.Fatalf("field was inappropriately applied to base metric")
	}
	if want, have := 1, len(c1.fields); want != have {
		t.Fatalf("field wasn't taken up by derived metric")
	}
}
