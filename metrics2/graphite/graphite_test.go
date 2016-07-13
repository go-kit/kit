package graphite

import (
	"bytes"
	"regexp"
	"strconv"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics2/teststat"
)

func TestCounter(t *testing.T) {
	prefix, name := "foo.", "bar"
	re := regexp.MustCompile(prefix + name + `.count ([0-9\.]+) [0-9]+`) // Graphite protocol
	g := NewRaw(prefix, log.NewNopLogger())
	counter := g.NewCounter(name)
	value := func() float64 {
		var buf bytes.Buffer
		g.WriteTo(&buf)
		match := re.FindStringSubmatch(buf.String())
		f, _ := strconv.ParseFloat(match[1], 64)
		return f
	}
	if err := teststat.TestCounter(counter, value); err != nil {
		t.Fatal(err)
	}
}

func TestGauge(t *testing.T) {
	prefix, name := "baz.", "quux"
	re := regexp.MustCompile(prefix + name + ` ([0-9\.]+) [0-9]+`)
	g := NewRaw(prefix, log.NewNopLogger())
	gauge := g.NewGauge(name)
	value := func() float64 {
		var buf bytes.Buffer
		g.WriteTo(&buf)
		match := re.FindStringSubmatch(buf.String())
		f, _ := strconv.ParseFloat(match[1], 64)
		return f
	}
	if err := teststat.TestGauge(gauge, value); err != nil {
		t.Fatal(err)
	}
}

func TestHistogram(t *testing.T) {
	prefix, name := "head.", "toes"
	re50 := regexp.MustCompile(prefix + name + `.p50 ([0-9\.]+) [0-9]+`)
	re90 := regexp.MustCompile(prefix + name + `.p90 ([0-9\.]+) [0-9]+`)
	re95 := regexp.MustCompile(prefix + name + `.p95 ([0-9\.]+) [0-9]+`)
	re99 := regexp.MustCompile(prefix + name + `.p99 ([0-9\.]+) [0-9]+`)
	g := NewRaw(prefix, log.NewNopLogger())
	histogram := g.NewHistogram(name, 50)
	quantiles := func() (float64, float64, float64, float64) {
		var buf bytes.Buffer
		g.WriteTo(&buf)
		match50 := re50.FindStringSubmatch(buf.String())
		p50, _ := strconv.ParseFloat(match50[1], 64)
		match90 := re90.FindStringSubmatch(buf.String())
		p90, _ := strconv.ParseFloat(match90[1], 64)
		match95 := re95.FindStringSubmatch(buf.String())
		p95, _ := strconv.ParseFloat(match95[1], 64)
		match99 := re99.FindStringSubmatch(buf.String())
		p99, _ := strconv.ParseFloat(match99[1], 64)
		return p50, p90, p95, p99
	}
	if err := teststat.TestHistogram(histogram, quantiles, 0.01); err != nil {
		t.Fatal(err)
	}
}

func TestWith(t *testing.T) {
	t.Skip("TODO")
}
