package prometheus_test

import (
	"strings"
	"testing"

	"github.com/peterbourgon/gokit/metrics"
	"github.com/peterbourgon/gokit/metrics/prometheus"
	"github.com/peterbourgon/gokit/metrics/teststat"
)

func TestPrometheusLabelBehavior(t *testing.T) {
	c := prometheus.NewCounter("test", "prometheus_label_behavior", "foobar", "Abc def.", []string{"used_key", "unused_key"})
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
	c := prometheus.NewCounter("test", "prometheus_counter", "foobar", "Lorem ipsum.", []string{})
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
	c := prometheus.NewGauge("test", "prometheus_gauge", "foobar", "Dolor sit.", []string{})
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

func TestPrometheusGaugeFloat(t *testing.T) {
	c := prometheus.NewGaugeFloat("test", "prometheus_gaugeFloat", "foobar", "Dolor sit.", []string{})
	c.Set(42.5)
	if want, have := strings.Join([]string{
		`# HELP test_prometheus_gaugeFloat_foobar Dolor sit.`,
		`# TYPE test_prometheus_gaugeFloat_foobar gauge`,
		`test_prometheus_gaugeFloat_foobar 42.5`,
	}, "\n"), teststat.ScrapePrometheus(t); !strings.Contains(have, want) {
		t.Errorf("metric stanza not found or incorrect\n%s", have)
	}
	c.Add(-43.5)
	if want, have := strings.Join([]string{
		`# HELP test_prometheus_gaugeFloat_foobar Dolor sit.`,
		`# TYPE test_prometheus_gaugeFloat_foobar gauge`,
		`test_prometheus_gaugeFloat_foobar -1`,
	}, "\n"), teststat.ScrapePrometheus(t); !strings.Contains(have, want) {
		t.Errorf("metric stanza not found or incorrect\n%s", have)
	}
}

func TestPrometheusHistogram(t *testing.T) {
	h := prometheus.NewHistogram("test", "prometheus_histogram", "foobar", "Qwerty asdf.", []string{})

	const mean, stdev int64 = 50, 10
	teststat.PopulateNormalHistogram(t, h, 34, mean, stdev)
	teststat.AssertPrometheusNormalHistogram(t, "test_prometheus_histogram_foobar", mean, stdev)
}
