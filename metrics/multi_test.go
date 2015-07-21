package metrics_test

import (
	stdexpvar "expvar"
	"fmt"
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

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/expvar"
	"github.com/go-kit/kit/metrics/prometheus"
)

func TestMultiWith(t *testing.T) {
	c := metrics.NewMultiCounter(
		expvar.NewCounter("foo"),
		prometheus.NewCounter(stdprometheus.CounterOpts{
			Namespace: "test",
			Subsystem: "multi_with",
			Name:      "bar",
			Help:      "Bar counter.",
		}, []string{"a"}),
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
		expvar.NewCounter("alpha"),
		prometheus.NewCounter(stdprometheus.CounterOpts{
			Namespace: "test",
			Subsystem: "multi_counter",
			Name:      "beta",
			Help:      "Beta counter.",
		}, []string{"a"}),
	).With(metrics.Field{Key: "a", Value: "b"}).Add(123)

	if want, have := "123", stdexpvar.Get("alpha").String(); want != have {
		t.Errorf("expvar: want %q, have %q", want, have)
	}

	if want, have := strings.Join([]string{
		`# HELP test_multi_counter_beta Beta counter.`,
		`# TYPE test_multi_counter_beta counter`,
		`test_multi_counter_beta{a="b"} 123`,
	}, "\n"), scrapePrometheus(t); !strings.Contains(have, want) {
		t.Errorf("Prometheus metric stanza not found or incorrect\n%s", have)
	}
}

func TestMultiGauge(t *testing.T) {
	g := metrics.NewMultiGauge(
		expvar.NewGauge("delta"),
		prometheus.NewGauge(stdprometheus.GaugeOpts{
			Namespace: "test",
			Subsystem: "multi_gauge",
			Name:      "kappa",
			Help:      "Kappa gauge.",
		}, []string{"a"}),
	)

	f := metrics.Field{Key: "a", Value: "aaa"}
	g.With(f).Set(34)

	if want, have := "34", stdexpvar.Get("delta").String(); want != have {
		t.Errorf("expvar: want %q, have %q", want, have)
	}
	if want, have := strings.Join([]string{
		`# HELP test_multi_gauge_kappa Kappa gauge.`,
		`# TYPE test_multi_gauge_kappa gauge`,
		`test_multi_gauge_kappa{a="aaa"} 34`,
	}, "\n"), scrapePrometheus(t); !strings.Contains(have, want) {
		t.Errorf("Prometheus metric stanza not found or incorrect\n%s", have)
	}

	g.With(f).Add(-40)

	if want, have := "-6", stdexpvar.Get("delta").String(); want != have {
		t.Errorf("expvar: want %q, have %q", want, have)
	}
	if want, have := strings.Join([]string{
		`# HELP test_multi_gauge_kappa Kappa gauge.`,
		`# TYPE test_multi_gauge_kappa gauge`,
		`test_multi_gauge_kappa{a="aaa"} -6`,
	}, "\n"), scrapePrometheus(t); !strings.Contains(have, want) {
		t.Errorf("Prometheus metric stanza not found or incorrect\n%s", have)
	}
}

func TestMultiHistogram(t *testing.T) {
	quantiles := []int{50, 90, 99}
	h := metrics.NewMultiHistogram(
		expvar.NewHistogram("omicron", 0, 100, 3, quantiles...),
		prometheus.NewSummary(stdprometheus.SummaryOpts{
			Namespace: "test",
			Subsystem: "multi_histogram",
			Name:      "nu",
			Help:      "Nu histogram.",
		}, []string{}),
	)

	const seed, mean, stdev int64 = 123, 50, 10
	populateNormalHistogram(t, h, seed, mean, stdev)
	assertExpvarNormalHistogram(t, "omicron", mean, stdev, quantiles)
	assertPrometheusNormalHistogram(t, `test_multi_histogram_nu`, mean, stdev)
}

func populateNormalHistogram(t *testing.T, h metrics.Histogram, seed int64, mean, stdev int64) {
	rand.Seed(seed)
	for i := 0; i < 1234; i++ {
		sample := int64(rand.NormFloat64()*float64(stdev) + float64(mean))
		h.Observe(sample)
	}
}

func assertExpvarNormalHistogram(t *testing.T, metricName string, mean, stdev int64, quantiles []int) {
	const tolerance int = 2
	for _, quantile := range quantiles {
		want := normalValueAtQuantile(mean, stdev, quantile)
		s := stdexpvar.Get(fmt.Sprintf("%s_p%02d", metricName, quantile)).String()
		have, err := strconv.Atoi(s)
		if err != nil {
			t.Fatal(err)
		}
		if int(math.Abs(float64(want)-float64(have))) > tolerance {
			t.Errorf("quantile %d: want %d, have %d", quantile, want, have)
		}
	}
}

func assertPrometheusNormalHistogram(t *testing.T, metricName string, mean, stdev int64) {
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
	re := name + `{quantile="` + quantileStr + `"} ([0-9]+)`
	matches := regexp.MustCompile(re).FindAllStringSubmatch(scrape, -1)
	if len(matches) < 1 {
		t.Fatalf("%q: quantile %q not found in scrape (%s)", name, quantileStr, re)
	}
	if len(matches[0]) < 2 {
		t.Fatalf("%q: quantile %q not found in scrape (%s)", name, quantileStr, re)
	}
	i, err := strconv.Atoi(matches[0][1])
	if err != nil {
		t.Fatal(err)
	}
	return i
}
