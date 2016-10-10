package influx

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	influxdb "github.com/influxdata/influxdb/client/v2"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/teststat"
)

func TestCounter(t *testing.T) {
	in := New(map[string]string{"a": "b"}, influxdb.BatchPointsConfig{}, log.NewNopLogger())
	re := regexp.MustCompile(`influx_counter,a=b count=([0-9\.]+) [0-9]+`) // reverse-engineered :\
	counter := in.NewCounter("influx_counter")
	value := func() float64 {
		client := &bufWriter{}
		in.WriteTo(client)
		match := re.FindStringSubmatch(client.buf.String())
		f, _ := strconv.ParseFloat(match[1], 64)
		return f
	}
	if err := teststat.TestCounter(counter, value); err != nil {
		t.Fatal(err)
	}
}

func ExampleCounter() {
	in := New(map[string]string{"a": "b"}, influxdb.BatchPointsConfig{}, log.NewNopLogger())
	counter := in.NewCounter("influx_counter")
	counter.Add(10)
	counter.With("error", "true").Add(1)
	counter.With("error", "false").Add(2)
	counter.Add(50)

	client := &bufWriter{}
	in.WriteTo(client)

	expectedLines := []string{
		`(influx_counter,a=b count=60) [0-9]{19}`,
		`(influx_counter,a=b,error=true count=1) [0-9]{19}`,
		`(influx_counter,a=b,error=false count=2) [0-9]{19}`,
	}

	if err := extractAndPrintMessage(expectedLines, client.buf.String()); err != nil {
		fmt.Println(err.Error())
	}

	// Output:
	// influx_counter,a=b count=60
	// influx_counter,a=b,error=true count=1
	// influx_counter,a=b,error=false count=2
}

func TestGauge(t *testing.T) {
	in := New(map[string]string{"foo": "alpha"}, influxdb.BatchPointsConfig{}, log.NewNopLogger())
	re := regexp.MustCompile(`influx_gauge,foo=alpha value=([0-9\.]+) [0-9]+`)
	gauge := in.NewGauge("influx_gauge")
	value := func() float64 {
		client := &bufWriter{}
		in.WriteTo(client)
		match := re.FindStringSubmatch(client.buf.String())
		f, _ := strconv.ParseFloat(match[1], 64)
		return f
	}
	if err := teststat.TestGauge(gauge, value); err != nil {
		t.Fatal(err)
	}
}

func ExampleGauge() {
	in := New(map[string]string{"a": "b"}, influxdb.BatchPointsConfig{}, log.NewNopLogger())
	gauge := in.NewGauge("influx_gauge")
	gauge.Set(10)
	gauge.With("error", "true").Set(2)
	gauge.With("error", "true").Set(1)
	gauge.With("error", "false").Set(2)
	gauge.Set(50)

	client := &bufWriter{}
	in.WriteTo(client)

	expectedLines := []string{
		`(influx_gauge,a=b value=50) [0-9]{19}`,
		`(influx_gauge,a=b,error=true value=1) [0-9]{19}`,
		`(influx_gauge,a=b,error=false value=2) [0-9]{19}`,
	}

	if err := extractAndPrintMessage(expectedLines, client.buf.String()); err != nil {
		fmt.Println(err.Error())
	}

	// Output:
	// influx_gauge,a=b value=50
	// influx_gauge,a=b,error=true value=1
	// influx_gauge,a=b,error=false value=2
}

func TestHistogram(t *testing.T) {
	in := New(map[string]string{"foo": "alpha"}, influxdb.BatchPointsConfig{}, log.NewNopLogger())
	re := regexp.MustCompile(`influx_histogram,bar=beta,foo=alpha p50=([0-9\.]+),p90=([0-9\.]+),p95=([0-9\.]+),p99=([0-9\.]+) [0-9]+`)
	histogram := in.NewHistogram("influx_histogram").With("bar", "beta")
	quantiles := func() (float64, float64, float64, float64) {
		w := &bufWriter{}
		in.WriteTo(w)
		match := re.FindStringSubmatch(w.buf.String())
		if len(match) != 5 {
			t.Errorf("These are not the quantiles you're looking for: %v\n", match)
		}
		var result [4]float64
		for i, q := range match[1:] {
			result[i], _ = strconv.ParseFloat(q, 64)
		}
		return result[0], result[1], result[2], result[3]
	}
	if err := teststat.TestHistogram(histogram, quantiles, 0.01); err != nil {
		t.Fatal(err)
	}
}

func ExampleHistogram() {
	in := New(map[string]string{"foo": "alpha"}, influxdb.BatchPointsConfig{}, log.NewNopLogger())
	histogram := in.NewHistogram("influx_histogram")
	histogram.Observe(float64(10))
	histogram.With("error", "true").Observe(float64(1))
	histogram.With("error", "false").Observe(float64(2))
	histogram.Observe(float64(50))

	client := &bufWriter{}
	in.WriteTo(client)

	expectedLines := []string{
		`(influx_histogram,foo=alpha p50=10,p90=50,p95=50,p99=50) [0-9]{19}`,
		`(influx_histogram,error=true,foo=alpha p50=1,p90=1,p95=1,p99=1) [0-9]{19}`,
		`(influx_histogram,error=false,foo=alpha p50=2,p90=2,p95=2,p99=2) [0-9]{19}`,
	}

	if err := extractAndPrintMessage(expectedLines, client.buf.String()); err != nil {
		fmt.Println(err.Error())
	}

	// Output:
	// influx_histogram,foo=alpha p50=10,p90=50,p95=50,p99=50
	// influx_histogram,error=true,foo=alpha p50=1,p90=1,p95=1,p99=1
	// influx_histogram,error=false,foo=alpha p50=2,p90=2,p95=2,p99=2
}

func TestHistogramLabels(t *testing.T) {
	in := New(map[string]string{}, influxdb.BatchPointsConfig{}, log.NewNopLogger())
	h := in.NewHistogram("foo")
	h.Observe(123)
	h.With("abc", "xyz").Observe(456)
	w := &bufWriter{}
	if err := in.WriteTo(w); err != nil {
		t.Fatal(err)
	}
	if want, have := 2, len(strings.Split(strings.TrimSpace(w.buf.String()), "\n")); want != have {
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

func extractAndPrintMessage(expected []string, msg string) error {
	for _, pattern := range expected {
		re := regexp.MustCompile(pattern)
		match := re.FindStringSubmatch(msg)
		if len(match) != 2 {
			return fmt.Errorf("Pattern not found! {%s} [%s]: %v\n", pattern, msg, match)
		}
		fmt.Println(match[1])
	}
	return nil
}
