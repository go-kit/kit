package influx

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics2/teststat"
	influxdb "github.com/influxdata/influxdb/client/v2"
)

func TestCounter(t *testing.T) {
	i := New(map[string]string{"a": "b"}, influxdb.BatchPointsConfig{}, log.NewNopLogger())
	re := regexp.MustCompile(`influx_counter,a=b,c=d count=([0-9\.]+) [0-9]+`) // reverse-engineered :\
	counter := i.NewCounter("influx_counter", map[string]string{"c": "d"})
	value := func() float64 {
		client := &bufWriter{}
		i.WriteTo(client)
		match := re.FindStringSubmatch(client.buf.String())
		f, _ := strconv.ParseFloat(match[1], 64)
		return f
	}
	if err := teststat.TestCounter(counter, value); err != nil {
		t.Fatal(err)
	}
}

func TestGauge(t *testing.T) {
	i := New(map[string]string{"foo": "alpha"}, influxdb.BatchPointsConfig{}, log.NewNopLogger())
	re := regexp.MustCompile(`influx_gauge,bar=beta,foo=alpha value=([0-9\.]+) [0-9]+`)
	gauge := i.NewGauge("influx_gauge", map[string]string{"bar": "beta"})
	value := func() float64 {
		client := &bufWriter{}
		i.WriteTo(client)
		match := re.FindStringSubmatch(client.buf.String())
		f, _ := strconv.ParseFloat(match[1], 64)
		return f
	}
	if err := teststat.TestGauge(gauge, value); err != nil {
		t.Fatal(err)
	}
}

func TestHistogram(t *testing.T) {
	i := New(map[string]string{"foo": "alpha"}, influxdb.BatchPointsConfig{}, log.NewNopLogger())
	re50 := regexp.MustCompile(`influx_histogram.p50,foo=alpha value=([0-9\.]+) [0-9]+`)
	re90 := regexp.MustCompile(`influx_histogram.p90,foo=alpha value=([0-9\.]+) [0-9]+`)
	re95 := regexp.MustCompile(`influx_histogram.p95,foo=alpha value=([0-9\.]+) [0-9]+`)
	re99 := regexp.MustCompile(`influx_histogram.p99,foo=alpha value=([0-9\.]+) [0-9]+`)
	histogram := i.NewHistogram("influx_histogram", map[string]string{}, 50)
	quantiles := func() (float64, float64, float64, float64) {
		w := &bufWriter{}
		i.WriteTo(w)
		match50 := re50.FindStringSubmatch(w.buf.String())
		p50, _ := strconv.ParseFloat(match50[1], 64)
		match90 := re90.FindStringSubmatch(w.buf.String())
		p90, _ := strconv.ParseFloat(match90[1], 64)
		match95 := re95.FindStringSubmatch(w.buf.String())
		p95, _ := strconv.ParseFloat(match95[1], 64)
		match99 := re99.FindStringSubmatch(w.buf.String())
		p99, _ := strconv.ParseFloat(match99[1], 64)
		return p50, p90, p95, p99
	}
	if err := teststat.TestHistogram(histogram, quantiles, 0.01); err != nil {
		t.Fatal(err)
	}
}

func TestHistogramLabels(t *testing.T) {
	i := New(map[string]string{}, influxdb.BatchPointsConfig{}, log.NewNopLogger())
	h := i.NewHistogram("foo", map[string]string{}, 50)
	h.Observe(123)
	h.With("abc", "xyz").Observe(456)

	w := &bufWriter{}
	if err := i.WriteTo(w); err != nil {
		t.Fatal(err)
	}
	fmt.Fprintf(os.Stderr, "%s\n", w.buf.String())

	if want, have := 2, len(strings.Split(w.buf.String(), "\n")); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

type bufWriter struct {
	buf bytes.Buffer
}

func (w *bufWriter) Write(bp influxdb.BatchPoints) error {
	for _, p := range bp.Points() {
		fmt.Fprintf(&w.buf, p.String()+"\n")
	}
	return nil
}
