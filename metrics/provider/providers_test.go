package provider

import (
	"testing"
	"time"

	"github.com/go-kit/kit/log"
)

func TestGraphite(t *testing.T) {
	p, err := NewGraphiteProvider("network", "address", "prefix", time.Second, log.NewNopLogger())
	if err != nil {
		t.Fatal(err)
	}
	testProvider(t, "Graphite", p)
}

func TestStatsd(t *testing.T) {
	p, err := NewStatsdProvider("network", "address", "prefix", time.Second, log.NewNopLogger())
	if err != nil {
		t.Fatal(err)
	}
	testProvider(t, "Statsd", p)
}

func TestDogStatsd(t *testing.T) {
	p, err := NewDogStatsdProvider("network", "address", "prefix", time.Second, log.NewNopLogger())
	if err != nil {
		t.Fatal(err)
	}
	testProvider(t, "DogStatsd", p)
}

func TestExpvar(t *testing.T) {
	testProvider(t, "Expvar", NewExpvarProvider("prefix"))
}

func TestPrometheus(t *testing.T) {
	testProvider(t, "Prometheus", NewPrometheusProvider("namespace", "subsystem"))
}

func testProvider(t *testing.T, what string, p Provider) {
	c := p.NewCounter("counter", "Counter help.")
	c.Add(1)

	h, err := p.NewHistogram("histogram", "Histogram help.", 1, 100, 3, 50, 95, 99)
	if err != nil {
		t.Errorf("%s: NewHistogram: %v", what, err)
	}
	h.Observe(99)

	g := p.NewGauge("gauge", "Gauge help.")
	g.Set(123)

	p.Stop()
}
