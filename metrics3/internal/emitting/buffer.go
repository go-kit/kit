package emitting

import (
	"fmt"
	"strings"
	"sync"

	"sort"

	"github.com/go-kit/kit/metrics3/generic"
)

type Buffer struct {
	buckets int

	mtx        sync.Mutex
	counters   map[point]*generic.Counter
	gauges     map[point]*generic.Gauge
	histograms map[point]*generic.Histogram
}

func (b *Buffer) Add(a Add) {
	pt := makePoint(a.Name, a.LabelValues)
	b.mtx.Lock()
	defer b.mtx.Unlock()
	c, ok := b.counters[pt]
	if !ok {
		c = generic.NewCounter(a.Name).With(a.LabelValues...).(*generic.Counter)
	}
	c.Add(a.Delta)
	b.counters[pt] = c
}

func (b *Buffer) Set(s Set) {
	pt := makePoint(s.Name, s.LabelValues)
	b.mtx.Lock()
	defer b.mtx.Unlock()
	g, ok := b.gauges[pt]
	if !ok {
		g = generic.NewGauge(s.Name).With(s.LabelValues...).(*generic.Gauge)
	}
	g.Set(s.Value)
	b.gauges[pt] = g
}

func (b *Buffer) Obv(o Obv) {
	pt := makePoint(o.Name, o.LabelValues)
	b.mtx.Lock()
	defer b.mtx.Unlock()
	h, ok := b.histograms[pt]
	if !ok {
		h = generic.NewHistogram(o.Name, b.buckets).With(o.LabelValues...).(*generic.Histogram)
	}
	h.Observe(o.Value)
	b.histograms[pt] = h
}

// point as in point in N-dimensional vector space;
// a string encoding of name + sorted k/v pairs.
type point string

const (
	recordDelimiter = "•"
	fieldDelimiter  = "·"
)

// (foo, [a b c d]) => "foo•a·b•c·d"
func makePoint(name string, labelValues []string) point {
	if len(labelValues)%2 != 0 {
		panic("odd number of label values; programmer error!")
	}
	pairs := make([]string, 0, len(labelValues)/2)
	for i := 0; i < len(labelValues); i += 2 {
		pairs = append(pairs, fmt.Sprintf("%s%s%s", labelValues[i], fieldDelimiter, labelValues[i+1]))
	}
	sort.Strings(sort.StringSlice(pairs))
	pairs = append([]string{name}, pairs...)
	return point(strings.Join(pairs, recordDelimiter))
}

// "foo•a·b•c·d" => (foo, [a b c d])
func (p point) nameLabelValues() (name string, labelValues []string) {
	records := strings.Split(string(p), recordDelimiter)
	if len(records)%2 != 1 { // always name + even number of label/values
		panic("even number of point records; programmer error!")
	}
	name, records = records[0], records[1:]
	labelValues = make([]string, 0, len(records)*2)
	for _, record := range records {
		fields := strings.SplitN(record, fieldDelimiter, 2)
		labelValues = append(labelValues, fields[0], fields[1])
	}
	return name, labelValues
}
