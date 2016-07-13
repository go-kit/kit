package prometheus

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

	"github.com/go-kit/kit/metrics2/teststat"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

func TestCounter(t *testing.T) {
	s := httptest.NewServer(stdprometheus.UninstrumentedHandler())
	defer s.Close()

	scrape := func() string {
		resp, _ := http.Get(s.URL)
		buf, _ := ioutil.ReadAll(resp.Body)
		return string(buf)
	}

	namespace, subsystem, name := "ns", "ss", "foo"
	re := regexp.MustCompile(namespace + `_` + subsystem + `_` + name + ` ([0-9\.]+)`)

	counter := NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
		Help:      "This is the help string.",
	}, []string{})

	value := func() float64 {
		matches := re.FindStringSubmatch(scrape())
		f, _ := strconv.ParseFloat(matches[1], 64)
		return f
	}

	if err := teststat.TestCounter(counter, value); err != nil {
		t.Fatal(err)
	}
}

func TestGauge(t *testing.T) {
	s := httptest.NewServer(stdprometheus.UninstrumentedHandler())
	defer s.Close()

	scrape := func() string {
		resp, _ := http.Get(s.URL)
		buf, _ := ioutil.ReadAll(resp.Body)
		return string(buf)
	}

	namespace, subsystem, name := "aaa", "bbb", "ccc"
	re := regexp.MustCompile(namespace + `_` + subsystem + `_` + name + ` ([0-9\.]+)`)

	gauge := NewGaugeFrom(stdprometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
		Help:      "This is a different help string.",
	}, []string{})

	value := func() float64 {
		matches := re.FindStringSubmatch(scrape())
		f, _ := strconv.ParseFloat(matches[1], 64)
		return f
	}

	if err := teststat.TestGauge(gauge, value); err != nil {
		t.Fatal(err)
	}
}

func TestSummary(t *testing.T) {
	s := httptest.NewServer(stdprometheus.UninstrumentedHandler())
	defer s.Close()

	scrape := func() string {
		resp, _ := http.Get(s.URL)
		buf, _ := ioutil.ReadAll(resp.Body)
		return string(buf)
	}

	namespace, subsystem, name := "test", "prometheus", "summary"
	re50 := regexp.MustCompile(namespace + `_` + subsystem + `_` + name + `{quantile="0.5"} ([0-9\.]+)`)
	re90 := regexp.MustCompile(namespace + `_` + subsystem + `_` + name + `{quantile="0.9"} ([0-9\.]+)`)
	re99 := regexp.MustCompile(namespace + `_` + subsystem + `_` + name + `{quantile="0.99"} ([0-9\.]+)`)

	summary := NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
		Help:      "This is the help string for the summary.",
	}, []string{})

	quantiles := func() (float64, float64, float64, float64) {
		buf := scrape()
		match50 := re50.FindStringSubmatch(buf)
		p50, _ := strconv.ParseFloat(match50[1], 64)
		match90 := re90.FindStringSubmatch(buf)
		p90, _ := strconv.ParseFloat(match90[1], 64)
		match99 := re99.FindStringSubmatch(buf)
		p99, _ := strconv.ParseFloat(match99[1], 64)
		p95 := p90 + ((p99 - p90) / 2) // Prometheus, y u no p95??? :< #yolo
		return p50, p90, p95, p99
	}

	if err := teststat.TestHistogram(summary, quantiles, 0.01); err != nil {
		t.Fatal(err)
	}
}

func TestHistogram(t *testing.T) {
	// Prometheus reports histograms as a count of observations that fell into
	// each predefined bucket, with the bucket value representing a global upper
	// limit. That is, the count monotonically increases over the buckets. This
	// requires a different strategy to test.

	s := httptest.NewServer(stdprometheus.UninstrumentedHandler())
	defer s.Close()

	scrape := func() string {
		resp, _ := http.Get(s.URL)
		buf, _ := ioutil.ReadAll(resp.Body)
		return string(buf)
	}

	namespace, subsystem, name := "test", "prometheus", "histogram"
	re := regexp.MustCompile(namespace + `_` + subsystem + `_` + name + `_bucket{le="([0-9]+|\+Inf)"} ([0-9\.]+)`)

	numStdev := 3
	bucketMin := (teststat.Mean - (numStdev * teststat.Stdev))
	bucketMax := (teststat.Mean + (numStdev * teststat.Stdev))
	if bucketMin < 0 {
		bucketMin = 0
	}
	bucketCount := 10
	bucketDelta := (bucketMax - bucketMin) / bucketCount
	buckets := []float64{}
	for i := bucketMin; i <= bucketMax; i += bucketDelta {
		buckets = append(buckets, float64(i))
	}

	histogram := NewHistogramFrom(stdprometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
		Help:      "This is the help string for the histogram.",
		Buckets:   buckets,
	}, []string{})

	// Can't TestHistogram, because Prometheus Histograms don't dynamically
	// compute quantiles. Instead, they fill up buckets. So, let's populate the
	// histogram kind of manually.
	teststat.PopulateNormalHistogram(histogram, rand.Int())

	// Then, we use ExpectedObservationsLessThan to validate.
	for _, line := range strings.Split(scrape(), "\n") {
		match := re.FindStringSubmatch(line)
		if match == nil {
			continue
		}

		bucket, _ := strconv.ParseInt(match[1], 10, 64)
		have, _ := strconv.ParseInt(match[2], 10, 64)

		want := teststat.ExpectedObservationsLessThan(bucket)
		if match[1] == "+Inf" {
			want = int64(teststat.Count) // special case
		}

		// Unfortunately, we observe experimentally that Prometheus is quite
		// imprecise at the extremes. I'm setting a very high tolerance for now.
		// It would be great to dig in and figure out whether that's a problem
		// with my Expected calculation, or in Prometheus.
		tolerance := 0.25
		if delta := math.Abs(float64(want) - float64(have)); (delta / float64(want)) > tolerance {
			t.Errorf("Bucket %d: want %d, have %d (%.1f%%)", bucket, want, have, (100.0 * delta / float64(want)))
		}
	}
}

func TestWith(t *testing.T) {
	t.Skip("TODO")
}
