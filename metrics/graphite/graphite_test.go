package graphite

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/teststat"
)

func TestHistogramQuantiles(t *testing.T) {
	prefix := "prefix"
	e := NewEmitter("", true, prefix, nil)
	var (
		name      = "test_histogram_quantiles"
		quantiles = []int{50, 90, 95, 99}
	)
	h, err := e.NewHistogram(name, 0, 100, 3, quantiles...)
	if err != nil {
		t.Fatalf("unable to create test histogram: ", err)
	}
	h = h.With(metrics.Field{Key: "ignored", Value: "field"})
	const seed, mean, stdev int64 = 424242, 50, 10
	teststat.PopulateNormalHistogram(t, h, seed, mean, stdev)

	// flush the current metrics into a buffer to examine
	var b bytes.Buffer
	e.flush(&b)
	teststat.AssertGraphiteNormalHistogram(t, prefix, name, mean, stdev, quantiles, b.String())
}

func TestCounter(t *testing.T) {
	var (
		prefix = "prefix"
		name   = "m"
		value  = 123
		e      = NewEmitter("", true, prefix, nil)
		b      bytes.Buffer
	)
	e.NewCounter(name).With(metrics.Field{Key: "ignored", Value: "field"}).Add(uint64(value))
	e.flush(&b)
	want := fmt.Sprintf("%s.%s.count %d", prefix, name, value)
	payload := b.String()
	if !strings.HasPrefix(payload, want) {
		t.Errorf("counter %s want\n%s, have\n%s", name, want, payload)
	}
}

func TestGauge(t *testing.T) {
	var (
		prefix = "prefix"
		name   = "xyz"
		value  = 54321
		delta  = 12345
		e      = NewEmitter("", true, prefix, nil)
		b      bytes.Buffer
		g      = e.NewGauge(name).With(metrics.Field{Key: "ignored", Value: "field"})
	)

	g.Set(float64(value))
	g.Add(float64(delta))

	e.flush(&b)
	payload := b.String()

	want := fmt.Sprintf("%s.%s %d", prefix, name, value+delta)
	if !strings.HasPrefix(payload, want) {
		t.Errorf("gauge %s want\n%s, have\n%s", name, want, payload)
	}
}
