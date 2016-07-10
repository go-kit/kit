package influx

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics2/teststat"
	influxdb "github.com/influxdata/influxdb/client/v2"
)

func TestCounter(t *testing.T) {
	i := NewRaw(map[string]string{"a": "b"}, influxdb.BatchPointsConfig{}, log.NewNopLogger())
	re := regexp.MustCompile(`influx_counter,a=b,c=d count=([0-9\.]+) [0-9]+`) // reverse-engineered :\
	counter := i.NewCounter("influx_counter", map[string]string{"c": "d"})
	value := func() float64 {
		client := &mockClient{}
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
	i := NewRaw(map[string]string{"foo": "alpha"}, influxdb.BatchPointsConfig{}, log.NewNopLogger())
	re := regexp.MustCompile(`influx_gauge,bar=beta,foo=alpha value=([0-9\.]+) [0-9]+`)
	gauge := i.NewGauge("influx_gauge", map[string]string{"bar": "beta"})
	value := func() float64 {
		client := &mockClient{}
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
	i := NewRaw(map[string]string{"foo": "alpha"}, influxdb.BatchPointsConfig{}, log.NewNopLogger())
	re50 := regexp.MustCompile(`influx_histogram.p50,foo=alpha value=([0-9\.]+) [0-9]+`)
	re90 := regexp.MustCompile(`influx_histogram.p90,foo=alpha value=([0-9\.]+) [0-9]+`)
	re95 := regexp.MustCompile(`influx_histogram.p95,foo=alpha value=([0-9\.]+) [0-9]+`)
	re99 := regexp.MustCompile(`influx_histogram.p99,foo=alpha value=([0-9\.]+) [0-9]+`)
	histogram := i.NewHistogram("influx_histogram", map[string]string{}, 50)
	quantiles := func() (float64, float64, float64, float64) {
		client := &mockClient{}
		i.WriteTo(client)
		match50 := re50.FindStringSubmatch(client.buf.String())
		p50, _ := strconv.ParseFloat(match50[1], 64)
		match90 := re90.FindStringSubmatch(client.buf.String())
		p90, _ := strconv.ParseFloat(match90[1], 64)
		match95 := re95.FindStringSubmatch(client.buf.String())
		p95, _ := strconv.ParseFloat(match95[1], 64)
		match99 := re99.FindStringSubmatch(client.buf.String())
		p99, _ := strconv.ParseFloat(match99[1], 64)
		return p50, p90, p95, p99
	}
	if err := teststat.TestHistogram(histogram, quantiles, 0.01); err != nil {
		t.Fatal(err)
	}
}

type mockClient struct {
	buf bytes.Buffer
}

func (c *mockClient) Write(bp influxdb.BatchPoints) error {
	for _, p := range bp.Points() {
		fmt.Fprintf(&c.buf, p.String()+"\n")
	}
	return nil
}

func (c *mockClient) Ping(time.Duration) (time.Duration, string, error) { return 1, "", nil }
func (c *mockClient) Query(influxdb.Query) (*influxdb.Response, error)  { return nil, nil }
func (c *mockClient) Close() error                                      { return nil }
