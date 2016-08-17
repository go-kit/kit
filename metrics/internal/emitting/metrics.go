package emitting

import (
	"github.com/go-kit/kit/metrics3"
	"github.com/go-kit/kit/metrics3/internal/lv"
)

type Counter struct {
	name       string
	lvs        lv.LabelValues
	sampleRate float64
	c          chan Add
}

type Add struct {
	Name        string
	LabelValues []string
	SampleRate  float64
	Delta       float64
}

func NewCounter(name string, sampleRate float64, c chan Add) *Counter {
	return &Counter{
		name:       name,
		sampleRate: sampleRate,
		c:          c,
	}
}

func (c *Counter) With(labelValues ...string) metrics.Counter {
	return &Counter{
		name:       c.name,
		lvs:        c.lvs.With(labelValues...),
		sampleRate: c.sampleRate,
		c:          c.c,
	}
}

func (c *Counter) Add(delta float64) {
	c.c <- Add{c.name, c.lvs, c.sampleRate, delta}
}

type Gauge struct {
	name string
	lvs  lv.LabelValues
	c    chan Set
}

type Set struct {
	Name        string
	LabelValues []string
	Value       float64
}

func NewGauge(name string, c chan Set) *Gauge {
	return &Gauge{
		name: name,
		c:    c,
	}
}

func (g *Gauge) With(labelValues ...string) metrics.Gauge {
	return &Gauge{
		name: g.name,
		lvs:  g.lvs.With(labelValues...),
		c:    g.c,
	}
}

func (g *Gauge) Set(value float64) {
	g.c <- Set{g.name, g.lvs, value}
}

type Histogram struct {
	name       string
	lvs        lv.LabelValues
	sampleRate float64
	c          chan Obv
}

type Obv struct {
	Name        string
	LabelValues []string
	SampleRate  float64
	Value       float64
}

func NewHistogram(name string, sampleRate float64, c chan Obv) *Histogram {
	return &Histogram{
		name:       name,
		sampleRate: sampleRate,
		c:          c,
	}
}

func (h *Histogram) With(labelValues ...string) metrics.Histogram {
	return &Histogram{
		name:       h.name,
		lvs:        h.lvs.With(labelValues...),
		sampleRate: h.sampleRate,
		c:          h.c,
	}
}

func (h *Histogram) Observe(value float64) {
	h.c <- Obv{h.name, h.lvs, h.sampleRate, value}
}
