package metrics_test

import (
	"expvar"
	"strings"
	"testing"

	"github.com/peterbourgon/gokit/metrics"
)

func TestMultiWith(t *testing.T) {
	c := metrics.NewMultiCounter(
		metrics.NewExpvarCounter("foo"),
		metrics.NewPrometheusCounter("test", "multi_with", "bar", "Bar counter.", []string{"a"}),
	)

	c.Add(1)
	c.With(metrics.Field{Key: "a", Value: "1"}).Add(2)
	c.Add(3)

	if want, have := strings.Join([]string{
		`# HELP test_multi_with_bar Bar counter.`,
		`# TYPE test_multi_with_bar counter`,
		`test_multi_with_bar{a="1"} 2`,
		`test_multi_with_bar{a="unknown"} 4`,
	}, "\n"), scrapePrometheus(t); !strings.Contains(have, want) {
		t.Errorf("Prometheus metric stanza not found or incorrect\n%s", have)
	}
}

func TestMultiCounter(t *testing.T) {
	metrics.NewMultiCounter(
		metrics.NewExpvarCounter("alpha"),
		metrics.NewPrometheusCounter("test", "multi_counter", "beta", "Beta counter.", []string{}),
	).Add(123)

	if want, have := "123", expvar.Get("alpha").String(); want != have {
		t.Errorf("expvar: want %q, have %q", want, have)
	}

	if want, have := strings.Join([]string{
		`# HELP test_multi_counter_beta Beta counter.`,
		`# TYPE test_multi_counter_beta counter`,
		`test_multi_counter_beta 123`,
	}, "\n"), scrapePrometheus(t); !strings.Contains(have, want) {
		t.Errorf("Prometheus metric stanza not found or incorrect\n%s", have)
	}
}

func TestMultiGauge(t *testing.T) {
	g := metrics.NewMultiGauge(
		metrics.NewExpvarGauge("delta"),
		metrics.NewPrometheusGauge("test", "multi_gauge", "kappa", "Kappa gauge.", []string{}),
	)

	g.Set(34)

	if want, have := "34", expvar.Get("delta").String(); want != have {
		t.Errorf("expvar: want %q, have %q", want, have)
	}
	if want, have := strings.Join([]string{
		`# HELP test_multi_gauge_kappa Kappa gauge.`,
		`# TYPE test_multi_gauge_kappa gauge`,
		`test_multi_gauge_kappa 34`,
	}, "\n"), scrapePrometheus(t); !strings.Contains(have, want) {
		t.Errorf("Prometheus metric stanza not found or incorrect\n%s", have)
	}

	g.Add(-40)

	if want, have := "-6", expvar.Get("delta").String(); want != have {
		t.Errorf("expvar: want %q, have %q", want, have)
	}
	if want, have := strings.Join([]string{
		`# HELP test_multi_gauge_kappa Kappa gauge.`,
		`# TYPE test_multi_gauge_kappa gauge`,
		`test_multi_gauge_kappa -6`,
	}, "\n"), scrapePrometheus(t); !strings.Contains(have, want) {
		t.Errorf("Prometheus metric stanza not found or incorrect\n%s", have)
	}
}

func TestMultiHistogram(t *testing.T) {
	quantiles := []int{50, 90, 99}
	h := metrics.NewMultiHistogram(
		metrics.NewExpvarHistogram("omicron", 0, 100, 3, quantiles...),
		metrics.NewPrometheusHistogram("test", "multi_histogram", "nu", "Nu histogram.", []string{}),
	)

	const seed, mean, stdev int64 = 123, 50, 10
	populateNormalHistogram(t, h, seed, mean, stdev)
	assertExpvarNormalHistogram(t, "omicron", mean, stdev, quantiles)
	assertPrometheusNormalHistogram(t, "test_multi_histogram_nu", mean, stdev)
}
