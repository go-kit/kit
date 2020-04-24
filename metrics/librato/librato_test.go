package librato

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/teststat"
)

const (
	metricNameToGenerateError         = "metric_name_used_to_throw_an_error"
	metricNameToGenerateInternalError = "metric_name_used_to_throw_an_internal_server_error"
)

type mockLibrato struct {
	*Librato
	valuesReceived *RequestPayload
	mtx            sync.RWMutex
}

type roundTripFunc func(r *http.Request) (*http.Response, error)

func (s roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return s(r)
}

func newMockLibrato() *mockLibrato {
	ml := &mockLibrato{
		valuesReceived: &RequestPayload{},
	}

	c := &http.Client{
		Timeout: 5 * time.Second,
		// mock HTTP Roundtrip to intercept what was send
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r == nil {
				return nil, nil
			}
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				return nil, err
			}

			vals := RequestPayload{}
			if err := json.Unmarshal(b, &vals); err != nil {
				return nil, err
			}

			if vals.Size() > 0 {
				ml.mtx.Lock()
				defer ml.mtx.Unlock()

				ml.valuesReceived.Counters = append(ml.valuesReceived.Counters, vals.Counters...)
				ml.valuesReceived.Gauges = append(ml.valuesReceived.Gauges, vals.Gauges...)

				// check for fake errors
				for _, vv := range ml.valuesReceived.Gauges {
					switch vv.Name {
					case metricNameToGenerateError:
						return nil, errors.New("tcp timeout")
					case metricNameToGenerateInternalError:
						return &http.Response{
							StatusCode: http.StatusInternalServerError,
							Body:       ioutil.NopCloser(bytes.NewReader([]byte("Internal Server Error!"))),
						}, nil
					}
				}

			}
			return &http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(bytes.NewReader(nil))}, nil
		}),
	}
	// WithLogger(log.NewNopLogger()),
	ml.Librato = New("", "", WithHttpClient(c))

	return ml
}

func TestCounter(t *testing.T) {
	name := "def"
	labels := []string{"source", "label"}
	lb := newMockLibrato()

	counter := lb.NewCounter(name).With(labels...)

	valuef := func() float64 {
		if err := lb.Send(); err != nil {
			t.Fatal(err)
		}
		lb.mtx.RLock()
		defer lb.mtx.RUnlock()
		c := lb.valuesReceived.Counters
		lb.valuesReceived.Counters = nil

		if l := len(c); l != 1 {
			t.Errorf("One counter expected, got %d", l)
			return 0
		}

		return *c[0].Value
	}
	if err := teststat.TestCounter(counter, valuef); err != nil {
		t.Fatal(err)
	}
	if err := teststat.TestCounter(counter, valuef); err != nil {
		t.Fatal("Fill and flush counter 2nd time: ", err)
	}
}

func TestCounterLowSendConcurrency(t *testing.T) {
	var names, labels, values []string
	for i := 1; i <= 45; i++ {
		num := strconv.Itoa(i)
		names = append(names, "name"+num)
		labels = append(labels, "source")
		values = append(values, num)
	}
	lb := newMockLibrato()
	lb.Librato.numConcurrentRequests = 2

	counters := make(map[string]metrics.Counter)
	var wants []float64
	for i, name := range names {
		counters[name] = lb.NewCounter(name).With(labels[i], values[i])
		wants = append(wants, teststat.FillCounter(counters[name]))
	}

	if err := lb.Send(); err != nil {
		t.Fatal(err)
	}

	metrics := lb.valuesReceived.Counters
	if len(metrics) != len(names) {
		t.Fatalf("Expected %d metrics, but got %d", len(names), len(metrics))
	}

	// sort metrics by source value
	sort.Slice(metrics, func(i, j int) bool {
		si, err := strconv.Atoi(*metrics[i].Source)
		if err != nil {
			t.Fatal(err)
		}
		sj, err := strconv.Atoi(*metrics[j].Source)
		if err != nil {
			t.Fatal(err)
		}
		return si < sj
	})

	for i, name := range names {
		m := metrics[i]
		if m.Name != name || m.Value == nil || *m.Value != wants[i] {
			t.Errorf("Expected metric %s=%f, got %s=%v", name, wants[i], m.Name, m.Value)
		}
	}
}

func TestAggregatedCounter(t *testing.T) {
	name := "agg_counter"
	lb := newMockLibrato()
	count := 45

	counter := lb.NewCounter(name).With("source", "test")
	var want float64

	for i := 1; i <= count; i++ {
		want += teststat.FillCounter(counter)
	}

	if err := lb.Send(); err != nil {
		t.Fatal(err)
	}

	metrics := lb.valuesReceived.Counters
	if len(metrics) != 1 {
		t.Fatalf("Expected 1 metrics, but got %d", len(metrics))
	}

	if m := metrics[0]; m.Value == nil || *m.Value != want {
		t.Fatalf("Expected value of %f, but got %v", want, m.Value)
	}
}

func TestCounterWithDifferentLabels(t *testing.T) {
	name := "test_counter"
	lb := newMockLibrato()
	count := 45

	counter := lb.NewCounter(name)
	var wants []float64

	for i := 1; i <= count; i++ {
		wants = append(wants, teststat.FillCounter(counter.With("source", strconv.Itoa(i))))
	}

	if err := lb.Send(); err != nil {
		t.Fatal(err)
	}

	metrics := lb.valuesReceived.Counters
	if len(metrics) != count {
		t.Fatalf("Expected %d metric, but got %d", count, len(metrics))
	}

	// sort metrics by source value
	sort.Slice(metrics, func(i, j int) bool {
		si, err := strconv.Atoi(*metrics[i].Source)
		if err != nil {
			t.Fatal(err)
		}
		sj, err := strconv.Atoi(*metrics[j].Source)
		if err != nil {
			t.Fatal(err)
		}
		return si < sj
	})

	for i, v := range wants {
		m := metrics[i]
		lbl := strconv.Itoa(i + 1)
		if m.Value == nil || *m.Value != v || *m.Source != lbl {
			t.Errorf("Expected metric source(%s)=%f, got source(%s)=%v", *m.Source, v, lbl, m.Value)
		}
	}
}

func TestHistogram(t *testing.T) {
	name := "def"
	labels := []string{"source", "value"}
	lb := newMockLibrato()

	histogram := lb.NewHistogram(name).With(labels...)
	n50 := fmt.Sprintf("%s_50", name)
	n90 := fmt.Sprintf("%s_90", name)
	n95 := fmt.Sprintf("%s_95", name)
	n99 := fmt.Sprintf("%s_99", name)

	quantiles := func() (p50, p90, p95, p99 float64) {
		err := lb.Send()
		if err != nil {
			t.Fatal(err)
		}

		metrics := lb.valuesReceived.Gauges
		lb.mtx.RLock()
		defer lb.mtx.RUnlock()

		for _, m := range metrics {
			if m.Name == n50 {
				p50 = *m.Value
			}
			if m.Name == n90 {
				p90 = *m.Value
			}
			if m.Name == n95 {
				p95 = *m.Value
			}
			if m.Name == n99 {
				p99 = *m.Value
			}
		}
		return
	}

	if err := teststat.TestHistogram(histogram, quantiles, 0.01); err != nil {
		t.Fatal(err)
	}

	// now test with only 2 custom percentiles
	//
	lb = newMockLibrato()
	lb.percentiles = []float64{0.50, 0.90}
	histogram = lb.NewHistogram(name).With(labels...)

	customQuantiles := func() (p50, p90, p95, p99 float64) {
		err := lb.Send()
		if err != nil {
			t.Fatal(err)
		}
		lb.mtx.RLock()
		defer lb.mtx.RUnlock()

		// our teststat.TestHistogram wants us to give p95 and p99,
		// but with custom percentiles we don't have those.
		// So fake them. Maybe we should make teststat.nvq() public and use that?
		p95 = 541.121341
		p99 = 558.158697

		metrics := lb.valuesReceived.Gauges
		for _, m := range metrics {
			if m.Name == n50 {
				p50 = *m.Value
			}
			if m.Name == n90 {
				p90 = *m.Value
			}

			// but fail if they are actually set (because that would mean the
			// WithPercentiles() is not respected)
			if m.Name == n95 {
				t.Fatal("p95 should not be set")
			}
			if m.Name == n99 {
				t.Fatal("p99 should not be set")
			}
		}

		return
	}
	if err := teststat.TestHistogram(histogram, customQuantiles, 0.01); err != nil {
		t.Fatal(err)
	}
}

func TestGauge(t *testing.T) {
	name := t.Name()
	lb := newMockLibrato()
	labels := []string{"source", "test"}
	avg := lb.NewGauge(name).With(labels...)
	count := 45

	var sum, sumSquares float64

	min := float64(1)
	max := float64(count)
	for v := min; v <= max; v++ {
		sum += v
		sumSquares += v * v
		avg.Add(v)
	}

	if err := lb.Send(); err != nil {
		t.Fatal(err)
	}

	metrics := lb.valuesReceived.Gauges
	if len(metrics) != 1 {
		t.Fatalf("Expected 1 metrics, but got %d", len(metrics))
	}

	want := GaugePayload{
		Metric: &Metric{
			Name:   name,
			Source: &labels[1],
		},
		Count:      &count,
		Min:        &min,
		Max:        &max,
		Sum:        &sum,
		SumSquares: &sumSquares,
	}

	if m := metrics[0]; !reflect.DeepEqual(m, want) {
		t.Fatalf("Expected\nvalue=nil, sum=%f, count=%d, max=%f, min=%f,\nbut got\nvalue=%v, sum=%v, count=%v, max=%v, min=%v",
			sum, count, max, min, m.Value, m.Sum, m.Count, m.Max, m.Min)
	}
}

func TestOneGauge(t *testing.T) {
	name := t.Name()
	labels := []string{"source", "label"}
	lb := newMockLibrato()
	v := 1.0
	lb.NewGauge(name).With(labels...).Add(v)

	if err := lb.Send(); err != nil {
		t.Fatal(err)
	}

	metrics := lb.valuesReceived.Gauges
	if len(metrics) != 1 {
		t.Fatalf("Expected 1 metrics, but got %d", len(metrics))
	}

	if m := metrics[0]; m.Value == nil || *m.Value != v {
		t.Fatalf("Expected value=%f but got value=%v", v, m.Value)
	}

}

func TestMetricsBatching(t *testing.T) {
	var names, labels, values []string
	for i := 1; i <= 250; i++ {
		num := strconv.Itoa(i)
		names = append(names, "name"+num)
		labels = append(labels, "source")
		values = append(values, num)
	}
	lb := newMockLibrato()

	counters := make(map[string]metrics.Counter)
	var wants []float64
	for i, name := range names {
		counters[name] = lb.NewCounter(name).With(labels[i], values[i])
		wants = append(wants, teststat.FillCounter(counters[name]))
	}

	if err := lb.Send(); err != nil {
		t.Fatal(err)
	}

	metrics := lb.valuesReceived.Counters
	if len(metrics) != len(names) {
		t.Fatalf("Expected %d metrics, but got %d", len(names), len(metrics))
	}

	// sort metrics by source value
	sort.Slice(metrics, func(i, j int) bool {
		si, err := strconv.Atoi(*metrics[i].Source)
		if err != nil {
			t.Fatal(err)
		}
		sj, err := strconv.Atoi(*metrics[j].Source)
		if err != nil {
			t.Fatal(err)
		}
		return si < sj
	})

	for i, name := range names {
		m := metrics[i]
		if m.Name != name || m.Value == nil || *m.Value != wants[i] {
			t.Errorf("Expected metric %s=%f, got %s=%v", name, wants[i], m.Name, m.Value)
		}
	}
}

func TestSumGauge(t *testing.T) {
	name := "sum_gauge"
	lb := newMockLibrato()

	g := lb.NewSumGauge(name).With("source", "test")
	count := 45

	var want float64

	for i := 1; i <= count; i++ {
		v := float64(i)
		want += v
		g.Add(v)
	}

	if err := lb.Send(); err != nil {
		t.Fatal(err)
	}

	metrics := lb.valuesReceived.Gauges
	if len(metrics) != 1 {
		t.Fatalf("Expected 1 metrics, but got %d", len(metrics))
	}

	if m := metrics[0]; m.Value == nil || *m.Value != want {
		t.Fatalf("Expected value of %f, but got %v", want, m.Value)
	}
}

func TestErrorLog(t *testing.T) {
	lb := newMockLibrato()

	lb.NewGauge(metricNameToGenerateError).Set(123)
	if err := lb.Send(); err == nil {
		t.Fatal("Expected error, but didn't get one")
	}

	lb.NewGauge(metricNameToGenerateInternalError).Set(123)
	if err := lb.Send(); err == nil {
		t.Fatal("Expected error, but didn't get one")
	}
}
