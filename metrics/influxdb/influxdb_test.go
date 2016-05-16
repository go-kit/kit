package influxdb_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/influxdb"
	stdinflux "github.com/influxdata/influxdb/client/v2"
)

func TestInfluxdbCounter(t *testing.T) {
	expectedName := "test_counter"
	expectedTags := map[string]string{}
	expectedFields := []map[string]interface{}{
		{"value": "2"},
		{"value": "7"},
		{"value": "10"},
	}

	cl := &mockClient{}
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

	for i := 0; i <= 2; i++ {
		givenPoint := mockPoint{
			Name:   expectedName,
			Tags:   expectedTags,
			Fields: expectedFields[i],
		}
		comparePoint(t, i, givenPoint, cl.Points[i])
	}
}

func TestInfluxdbCounterWithTag(t *testing.T) {
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
	givenName := given.Name()
	if givenName != expected.Name {
		t.Errorf("Point %v invalid name, expected %v got %v", i, expected.Name, givenName)
	}

	tagsEqual := reflect.DeepEqual(expected.Tags, given.Tags())
	if !tagsEqual {
		t.Errorf("Point %v invalid tags, expected %v got %v", i, expected.Tags, given.Tags())
	}

	fieldsEqual := reflect.DeepEqual(expected.Fields, given.Fields())
	if !fieldsEqual {
		t.Errorf("Point %v invalid fields, expected %v got %v", i, expected.Fields, given.Fields())
	}
}

type mockClient struct {
	Points []stdinflux.Point
}

func (m *mockClient) Ping(timeout time.Duration) (time.Duration, string, error) {
	t := 0 * time.Millisecond
	return t, "", nil
}

func (m *mockClient) Write(bp stdinflux.BatchPoints) error {
	for _, p := range bp.Points() {
		m.Points = append(m.Points, *p)
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
