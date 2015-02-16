package prometheus_test

import (
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"

	stdprometheus "github.com/prometheus/client_golang/prometheus"

	"github.com/peterbourgon/gokit/metrics"
	"github.com/peterbourgon/gokit/metrics/prometheus"
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
	}, "\n"), scrapePrometheus(t); !strings.Contains(have, want) {
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
	}, "\n"), scrapePrometheus(t); !strings.Contains(have, want) {
		t.Errorf("metric stanza not found or incorrect\n%s", have)
	}
	c.Add(3)
	c.Add(4)
	if want, have := strings.Join([]string{
		`# HELP test_prometheus_counter_foobar Lorem ipsum.`,
		`# TYPE test_prometheus_counter_foobar counter`,
		`test_prometheus_counter_foobar 10`,
	}, "\n"), scrapePrometheus(t); !strings.Contains(have, want) {
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
	}, "\n"), scrapePrometheus(t); !strings.Contains(have, want) {
		t.Errorf("metric stanza not found or incorrect\n%s", have)
	}
	c.Add(-43)
	if want, have := strings.Join([]string{
		`# HELP test_prometheus_gauge_foobar Dolor sit.`,
		`# TYPE test_prometheus_gauge_foobar gauge`,
		`test_prometheus_gauge_foobar -1`,
	}, "\n"), scrapePrometheus(t); !strings.Contains(have, want) {
		t.Errorf("metric stanza not found or incorrect\n%s", have)
	}
}

func TestPrometheusHistogram(t *testing.T) {
	h := prometheus.NewHistogram("test", "prometheus_histogram", "foobar", "Qwerty asdf.", []string{})

	const mean, stdev int64 = 50, 10
	populateNormalHistogram(t, h, 34, mean, stdev)
	assertNormalHistogram(t, "test_prometheus_histogram_foobar", mean, stdev)
}

func populateNormalHistogram(t *testing.T, h metrics.Histogram, seed int64, mean, stdev int64) {
	rand.Seed(seed)
	for i := 0; i < 1234; i++ {
		sample := int64(rand.NormFloat64()*float64(stdev) + float64(mean))
		h.Observe(sample)
	}
}

func assertNormalHistogram(t *testing.T, metricName string, mean, stdev int64) {
	scrape := scrapePrometheus(t)
	const tolerance int = 5 // Prometheus approximates higher quantiles badly -_-;
	for quantileInt, quantileStr := range map[int]string{50: "0.5", 90: "0.9", 99: "0.99"} {
		want := normalValueAtQuantile(mean, stdev, quantileInt)
		have := getPrometheusQuantile(t, scrape, metricName, quantileStr)
		if int(math.Abs(float64(want)-float64(have))) > tolerance {
			t.Errorf("%q: want %d, have %d", quantileStr, want, have)
		}
	}
}

// https://en.wikipedia.org/wiki/Normal_distribution#Quantile_function
func normalValueAtQuantile(mean, stdev int64, quantile int) int64 {
	return int64(float64(mean) + float64(stdev)*math.Sqrt2*erfinv(2*(float64(quantile)/100)-1))
}

// https://stackoverflow.com/questions/5971830/need-code-for-inverse-error-function
func erfinv(y float64) float64 {
	if y < -1.0 || y > 1.0 {
		panic("invalid input")
	}

	var (
		a = [4]float64{0.886226899, -1.645349621, 0.914624893, -0.140543331}
		b = [4]float64{-2.118377725, 1.442710462, -0.329097515, 0.012229801}
		c = [4]float64{-1.970840454, -1.624906493, 3.429567803, 1.641345311}
		d = [2]float64{3.543889200, 1.637067800}
	)

	const y0 = 0.7
	var x, z float64

	if math.Abs(y) == 1.0 {
		x = -y * math.Log(0.0)
	} else if y < -y0 {
		z = math.Sqrt(-math.Log((1.0 + y) / 2.0))
		x = -(((c[3]*z+c[2])*z+c[1])*z + c[0]) / ((d[1]*z+d[0])*z + 1.0)
	} else {
		if y < y0 {
			z = y * y
			x = y * (((a[3]*z+a[2])*z+a[1])*z + a[0]) / ((((b[3]*z+b[3])*z+b[1])*z+b[0])*z + 1.0)
		} else {
			z = math.Sqrt(-math.Log((1.0 - y) / 2.0))
			x = (((c[3]*z+c[2])*z+c[1])*z + c[0]) / ((d[1]*z+d[0])*z + 1.0)
		}
		x = x - (math.Erf(x)-y)/(2.0/math.SqrtPi*math.Exp(-x*x))
		x = x - (math.Erf(x)-y)/(2.0/math.SqrtPi*math.Exp(-x*x))
	}

	return x
}

func scrapePrometheus(t *testing.T) string {
	server := httptest.NewServer(stdprometheus.UninstrumentedHandler())
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	return strings.TrimSpace(string(buf))
}

func getPrometheusQuantile(t *testing.T, scrape, name, quantileStr string) int {
	matches := regexp.MustCompile(name+`{quantile="`+quantileStr+`"} ([0-9]+)`).FindAllStringSubmatch(scrape, -1)
	if len(matches) < 1 {
		t.Fatalf("%q: quantile %q not found in scrape", name, quantileStr)
	}
	if len(matches[0]) < 2 {
		t.Fatalf("%q: quantile %q not found in scrape", name, quantileStr)
	}
	i, err := strconv.Atoi(matches[0][1])
	if err != nil {
		t.Fatal(err)
	}
	return i
}
