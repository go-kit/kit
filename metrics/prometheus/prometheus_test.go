package prometheus_test

import (
	"strings"
	"testing"

	stdprometheus "github.com/prometheus/client_golang/prometheus"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	"github.com/go-kit/kit/metrics/teststat"
)

func TestPrometheusLabelBehavior(t *testing.T) {
	c := prometheus.NewCounter(stdprometheus.CounterOpts{
		Namespace: "test",
		Subsystem: "prometheus_label_behavior",
		Name:      "foobar",
		Help:      "Abc def.",
	}, []string{"used_key", "unused_key"})
	c.With(metrics.Field{Key: "used_key", Value: "declared"}).Add(1)
	c.Add(1)

	if want, have := strings.Join([]string{
		`# HELP test_prometheus_label_behavior_foobar Abc def.`,
		`# TYPE test_prometheus_label_behavior_foobar counter`,
		`test_prometheus_label_behavior_foobar{unused_key="unknown",used_key="declared"} 1`,
		`test_prometheus_label_behavior_foobar{unused_key="unknown",used_key="unknown"} 1`,
	}, "\n"), teststat.ScrapePrometheus(t); !strings.Contains(have, want) {
		t.Errorf("metric stanza not found or incorrect\n%s", have)
	}
}

func TestPrometheusCounter(t *testing.T) {
	c := prometheus.NewCounter(stdprometheus.CounterOpts{
		Namespace: "test",
		Subsystem: "prometheus_counter",
		Name:      "foobar",
		Help:      "Lorem ipsum.",
	}, []string{})
	c.Add(1)
	c.Add(2)
	if want, have := strings.Join([]string{
		`# HELP test_prometheus_counter_foobar Lorem ipsum.`,
		`# TYPE test_prometheus_counter_foobar counter`,
		`test_prometheus_counter_foobar 3`,
	}, "\n"), teststat.ScrapePrometheus(t); !strings.Contains(have, want) {
		t.Errorf("metric stanza not found or incorrect\n%s", have)
	}
	c.Add(3)
	c.Add(4)
	if want, have := strings.Join([]string{
		`# HELP test_prometheus_counter_foobar Lorem ipsum.`,
		`# TYPE test_prometheus_counter_foobar counter`,
		`test_prometheus_counter_foobar 10`,
	}, "\n"), teststat.ScrapePrometheus(t); !strings.Contains(have, want) {
		t.Errorf("metric stanza not found or incorrect\n%s", have)
	}
}

func TestPrometheusGauge(t *testing.T) {
	c := prometheus.NewGauge(stdprometheus.GaugeOpts{
		Namespace: "test",
		Subsystem: "prometheus_gauge",
		Name:      "foobar",
		Help:      "Dolor sit.",
	}, []string{})
	c.Set(42)
	if want, have := strings.Join([]string{
		`# HELP test_prometheus_gauge_foobar Dolor sit.`,
		`# TYPE test_prometheus_gauge_foobar gauge`,
		`test_prometheus_gauge_foobar 42`,
	}, "\n"), teststat.ScrapePrometheus(t); !strings.Contains(have, want) {
		t.Errorf("metric stanza not found or incorrect\n%s", have)
	}
	c.Add(-43)
	if want, have := strings.Join([]string{
		`# HELP test_prometheus_gauge_foobar Dolor sit.`,
		`# TYPE test_prometheus_gauge_foobar gauge`,
		`test_prometheus_gauge_foobar -1`,
	}, "\n"), teststat.ScrapePrometheus(t); !strings.Contains(have, want) {
		t.Errorf("metric stanza not found or incorrect\n%s", have)
	}
}

func TestPrometheusCallbackGauge(t *testing.T) {
	value := 123.456
	cb := func() float64 { return value }
	prometheus.RegisterCallbackGauge(stdprometheus.GaugeOpts{
		Namespace: "test",
		Subsystem: "prometheus_gauge",
		Name:      "bazbaz",
		Help:      "Help string.",
	}, cb)
	if want, have := strings.Join([]string{
		`# HELP test_prometheus_gauge_bazbaz Help string.`,
		`# TYPE test_prometheus_gauge_bazbaz gauge`,
		`test_prometheus_gauge_bazbaz 123.456`,
	}, "\n"), teststat.ScrapePrometheus(t); !strings.Contains(have, want) {
		t.Errorf("metric stanza not found or incorrect\n%s", have)
	}
}

func TestPrometheusSummary(t *testing.T) {
	h := prometheus.NewSummary(stdprometheus.SummaryOpts{
		Namespace: "test",
		Subsystem: "prometheus_summary_histogram",
		Name:      "foobar",
		Help:      "Qwerty asdf.",
	}, []string{})

	const mean, stdev int64 = 50, 10
	teststat.PopulateNormalHistogram(t, h, 34, mean, stdev)
	teststat.AssertPrometheusNormalSummary(t, "test_prometheus_summary_histogram_foobar", mean, stdev)
}

func TestPrometheusHistogram(t *testing.T) {
	buckets := []float64{20, 40, 60, 80, 100}
	h := prometheus.NewHistogram(stdprometheus.HistogramOpts{
		Namespace: "test",
		Subsystem: "prometheus_histogram_histogram",
		Name:      "quux",
		Help:      "Qwerty asdf.",
		Buckets:   buckets,
	}, []string{})

	const mean, stdev int64 = 50, 10
	teststat.PopulateNormalHistogram(t, h, 34, mean, stdev)
	teststat.AssertPrometheusBucketedHistogram(t, "test_prometheus_histogram_histogram_quux_bucket", mean, stdev, buckets)
}
