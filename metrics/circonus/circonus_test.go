package circonus

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/circonus-labs/circonus-gometrics"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/teststat"
)

var (
	// The Circonus Start() method launches a new goroutine that cannot be
	// stopped. So, make sure we only do that once per test run.
	onceStart sync.Once

	// Similarly, once set, the submission interval cannot be changed.
	submissionInterval = 50 * time.Millisecond
)

func TestCounter(t *testing.T) {
	log.SetOutput(ioutil.Discard)   // Circonus logs errors directly! Bad Circonus!
	defer circonusgometrics.Reset() // Circonus has package global state! Bad Circonus!

	var (
		name  = "test_counter"
		value uint64
		mtx   sync.Mutex
	)
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type postCounter struct {
			Value uint64 `json:"_value"` // reverse-engineered
		}
		m := map[string]postCounter{}
		json.NewDecoder(r.Body).Decode(&m)
		mtx.Lock()
		defer mtx.Unlock()
		value = m[name].Value
	}))
	defer s.Close()

	// We must set the submission URL before making observations. Circonus emits
	// using a package-global goroutine that is started with Start() but not
	// stoppable. Once it gets going, it will POST to the package-global
	// submission URL every interval. If we record observations first, and then
	// try to change the submission URL, it's possible that the observations
	// have already been submitted to the previous URL. And, at least in the
	// case of histograms, every submit, success or failure, resets the data.
	// Bad Circonus!

	circonusgometrics.WithSubmissionUrl(s.URL)
	circonusgometrics.WithInterval(submissionInterval)

	c := NewCounter(name)

	if want, have := name, c.Name(); want != have {
		t.Errorf("want %q, have %q", want, have)
	}

	c.Add(123)
	c.With(metrics.Field{Key: "this should", Value: "be ignored"}).Add(456)

	onceStart.Do(func() { circonusgometrics.Start() })
	if err := within(time.Second, func() bool {
		mtx.Lock()
		defer mtx.Unlock()
		return value > 0
	}); err != nil {
		t.Fatalf("error collecting results: %v", err)
	}

	if want, have := 123+456, int(value); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestGauge(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	defer circonusgometrics.Reset()

	var (
		name  = "test_gauge"
		value float64
		mtx   sync.Mutex
	)
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type postGauge struct {
			Value float64 `json:"_value"`
		}
		m := map[string]postGauge{}
		json.NewDecoder(r.Body).Decode(&m)
		mtx.Lock()
		defer mtx.Unlock()
		value = m[name].Value
	}))
	defer s.Close()

	circonusgometrics.WithSubmissionUrl(s.URL)
	circonusgometrics.WithInterval(submissionInterval)

	g := NewGauge(name)

	g.Set(123)
	g.Add(456) // is a no-op

	if want, have := 0.0, g.Get(); want != have {
		t.Errorf("Get should always return %.2f, but I got %.2f", want, have)
	}

	onceStart.Do(func() { circonusgometrics.Start() })

	if err := within(time.Second, func() bool {
		mtx.Lock()
		defer mtx.Unlock()
		return value > 0.0
	}); err != nil {
		t.Fatalf("error collecting results: %v", err)
	}

	if want, have := 123.0, value; want != have {
		t.Errorf("want %.2f, have %.2f", want, have)
	}
}

func TestHistogram(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	defer circonusgometrics.Reset()

	var (
		name       = "test_histogram"
		result     []string
		mtx        sync.Mutex
		onceDecode sync.Once
	)
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type postHistogram struct {
			Value []string `json:"_value"`
		}
		onceDecode.Do(func() {
			m := map[string]postHistogram{}
			json.NewDecoder(r.Body).Decode(&m)
			mtx.Lock()
			defer mtx.Unlock()
			result = m[name].Value
		})
	}))
	defer s.Close()

	circonusgometrics.WithSubmissionUrl(s.URL)
	circonusgometrics.WithInterval(submissionInterval)

	h := NewHistogram(name)

	var (
		seed  = int64(123)
		mean  = int64(500)
		stdev = int64(123)
		min   = int64(0)
		max   = 2 * mean
	)
	teststat.PopulateNormalHistogram(t, h, seed, mean, stdev)

	onceStart.Do(func() { circonusgometrics.Start() })

	if err := within(time.Second, func() bool {
		mtx.Lock()
		defer mtx.Unlock()
		return len(result) > 0
	}); err != nil {
		t.Fatalf("error collecting results: %v", err)
	}

	teststat.AssertCirconusNormalHistogram(t, mean, stdev, min, max, result)
}

func within(d time.Duration, f func() bool) error {
	deadline := time.Now().Add(d)
	for {
		if time.Now().After(deadline) {
			return errors.New("deadline exceeded")
		}
		if f() {
			return nil
		}
		time.Sleep(d / 10)
	}
}
