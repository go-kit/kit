package influxdb_test

import (
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/influxdb"
	stdinflux "github.com/influxdata/influxdb/client/v2"
)

func TestCounter(t *testing.T) {
	expectedName := "test_counter"
	expectedTags := map[string]string{}
	expectedFields := []map[string]interface{}{
		{"value": "2"},
		{"value": "7"},
		{"value": "10"},
	}

	cl := &mockClient{}
	cl.Add(3)
	bp, _ := stdinflux.NewBatchPoints(stdinflux.BatchPointsConfig{
		Database:  "testing",
		Precision: "s",
	})

	tags := []metrics.Field{}
	for key, value := range expectedTags {
		tags = append(tags, metrics.Field{Key: key, Value: value})
	}

	triggerChan := make(chan time.Time)
	counter := influxdb.NewCounterTick(cl, bp, expectedName, tags, triggerChan)
	counter.Add(2)
	counter.Add(5)
	counter.Add(3)

	triggerChan <- time.Now()
	cl.Wait()

	for i := 0; i <= 2; i++ {
		givenPoint := mockPoint{
			Name:   expectedName,
			Tags:   expectedTags,
			Fields: expectedFields[i],
		}
		comparePoint(t, i, givenPoint, cl.Points[i])
	}
}

func TestCounterWithTags(t *testing.T) {
	expectedName := "test_counter"
	expectedTags := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	expectedFields := []map[string]interface{}{
		{"value": "2"},
		{"Test": "Test", "value": "7"},
		{"Test": "Test", "value": "10"},
	}

	cl := &mockClient{}
	cl.Add(3)
	bp, _ := stdinflux.NewBatchPoints(stdinflux.BatchPointsConfig{
		Database:  "testing",
		Precision: "s",
	})

	tags := []metrics.Field{}
	for key, value := range expectedTags {
		tags = append(tags, metrics.Field{Key: key, Value: value})
	}

	triggerChan := make(chan time.Time)
	counter := influxdb.NewCounterTick(cl, bp, expectedName, tags, triggerChan)
	counter.Add(2)
	counter = counter.With(metrics.Field{Key: "Test", Value: "Test"})
	counter.Add(5)
	counter.Add(3)

	triggerChan <- time.Now()
	cl.Wait()

	for i := 0; i <= 2; i++ {
		givenPoint := mockPoint{
			Name:   expectedName,
			Tags:   expectedTags,
			Fields: expectedFields[i],
		}
		comparePoint(t, i, givenPoint, cl.Points[i])
	}
}

func comparePoint(t *testing.T, i int, expected mockPoint, given stdinflux.Point) {

	if want, have := expected.Name, given.Name(); want != have {
		t.Errorf("point %d: want %q, have %q", i, want, have)
	}

	if want, have := expected.Tags, given.Tags(); !reflect.DeepEqual(want, have) {
		t.Errorf("point %d: want %v, have %v", i, want, have)
	}

	if want, have := expected.Fields, given.Fields(); !reflect.DeepEqual(want, have) {
		t.Errorf("point %d: want %v, have %v", i, want, have)
	}
}

type mockClient struct {
	Points []stdinflux.Point
	sync.WaitGroup
}

func (m *mockClient) Ping(timeout time.Duration) (time.Duration, string, error) {
	t := 0 * time.Millisecond
	return t, "", nil
}

func (m *mockClient) Write(bp stdinflux.BatchPoints) error {
	for _, p := range bp.Points() {
		m.Points = append(m.Points, *p)
		m.Done()
	}

	return nil
}

func (m *mockClient) Query(q stdinflux.Query) (*stdinflux.Response, error) {
	return nil, nil
}

func (m *mockClient) Close() error {
	return nil
}

type mockPoint struct {
	Name   string
	Tags   map[string]string
	Fields map[string]interface{}
}
