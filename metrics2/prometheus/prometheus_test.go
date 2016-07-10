package prometheus

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
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
