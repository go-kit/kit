package teststat

import (
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

// ScrapePrometheus returns the text encoding of the current state of
// Prometheus.
func ScrapePrometheus(t *testing.T) string {
	server := httptest.NewServer(prometheus.UninstrumentedHandler())
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

// AssertPrometheusNormalSummary ensures the Prometheus Summary referenced by
// name abides a normal distribution.
func AssertPrometheusNormalSummary(t *testing.T, metricName string, mean, stdev int64) {
	scrape := ScrapePrometheus(t)
	const tolerance int = 5 // Prometheus approximates higher quantiles badly -_-;
	for quantileInt, quantileStr := range map[int]string{50: "0.5", 90: "0.9", 99: "0.99"} {
		want := normalValueAtQuantile(mean, stdev, quantileInt)
		have := getPrometheusQuantile(t, scrape, metricName, quantileStr)
		if int(math.Abs(float64(want)-float64(have))) > tolerance {
			t.Errorf("%q: want %d, have %d", quantileStr, want, have)
		}
	}
}

// AssertPrometheusBucketedHistogram ensures the Prometheus Histogram
// referenced by name has observations in the expected quantity and bucket.
func AssertPrometheusBucketedHistogram(t *testing.T, metricName string, mean, stdev int64, buckets []float64) {
	scrape := ScrapePrometheus(t)
	const tolerance int = population / 50 // pretty coarse-grained
	for _, bucket := range buckets {
		want := observationsLessThan(mean, stdev, bucket, population)
		have := getPrometheusLessThan(t, scrape, metricName, strconv.FormatFloat(bucket, 'f', 0, 64))
		if int(math.Abs(float64(want)-float64(have))) > tolerance {
			t.Errorf("%.0f: want %d, have %d", bucket, want, have)
		}
	}
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

func getPrometheusLessThan(t *testing.T, scrape, name, target string) int {
	matches := regexp.MustCompile(name+`{le="`+target+`"} ([0-9]+)`).FindAllStringSubmatch(scrape, -1)
	if len(matches) < 1 {
		t.Logf(">>>\n%s\n", scrape)
		t.Fatalf("%q: bucket %q not found in scrape", name, target)
	}
	if len(matches[0]) < 2 {
		t.Fatalf("%q: bucket %q not found in scrape", name, target)
	}
	i, err := strconv.Atoi(matches[0][1])
	if err != nil {
		t.Fatal(err)
	}
	return i
}
