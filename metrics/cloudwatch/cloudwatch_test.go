package cloudwatch

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/teststat"
)

type mockCloudWatch struct {
	cloudwatchiface.CloudWatchAPI
	mtx                sync.RWMutex
	valuesReceived     map[string]float64
	dimensionsReceived map[string][]*cloudwatch.Dimension
}

func newMockCloudWatch() *mockCloudWatch {
	return &mockCloudWatch{
		valuesReceived:     map[string]float64{},
		dimensionsReceived: map[string][]*cloudwatch.Dimension{},
	}
}

func (mcw *mockCloudWatch) PutMetricData(input *cloudwatch.PutMetricDataInput) (*cloudwatch.PutMetricDataOutput, error) {
	mcw.mtx.Lock()
	defer mcw.mtx.Unlock()
	for _, datum := range input.MetricData {
		mcw.valuesReceived[*datum.MetricName] = *datum.Value
		mcw.dimensionsReceived[*datum.MetricName] = datum.Dimensions
	}
	return nil, nil
}

func testDimensions(svc *mockCloudWatch, name string, labelValues ...string) error {
	dimensions, ok := svc.dimensionsReceived[name]
	if !ok {
		if len(labelValues) > 0 {
			return errors.New("Expected dimensions to be available, but none were")
		}
	}
LabelValues:
	for i, j := 0, 0; i < len(labelValues); i, j = i+2, j+1 {
		name, value := labelValues[i], labelValues[i+1]
		for _, dimension := range dimensions {
			if *dimension.Name == name {
				if *dimension.Value == value {
					break LabelValues
				}
			}
		}
		return fmt.Errorf("Could not find dimension with name %s and value %s", name, value)
	}

	return nil
}

func TestCounter(t *testing.T) {
	namespace, name := "abc", "def"
	label, value := "label", "value"
	svc := newMockCloudWatch()
	cw := New(namespace, svc, log.NewNopLogger())
	counter := cw.NewCounter(name).With(label, value)
	valuef := func() float64 {
		err := cw.Send()
		if err != nil {
			t.Fatal(err)
		}
		svc.mtx.RLock()
		defer svc.mtx.RUnlock()
		return svc.valuesReceived[name]
	}
	if err := teststat.TestCounter(counter, valuef); err != nil {
		t.Fatal(err)
	}
	if err := testDimensions(svc, name, label, value); err != nil {
		t.Fatal(err)
	}
}

func TestGauge(t *testing.T) {
	namespace, name := "abc", "def"
	label, value := "label", "value"
	svc := newMockCloudWatch()
	cw := New(namespace, svc, log.NewNopLogger())
	gauge := cw.NewGauge(name).With(label, value)
	valuef := func() float64 {
		err := cw.Send()
		if err != nil {
			t.Fatal(err)
		}
		svc.mtx.RLock()
		defer svc.mtx.RUnlock()
		return svc.valuesReceived[name]
	}
	if err := teststat.TestGauge(gauge, valuef); err != nil {
		t.Fatal(err)
	}
	if err := testDimensions(svc, name, label, value); err != nil {
		t.Fatal(err)
	}
}

func TestHistogram(t *testing.T) {
	namespace, name := "abc", "def"
	label, value := "label", "value"
	svc := newMockCloudWatch()
	cw := New(namespace, svc, log.NewNopLogger())
	histogram := cw.NewHistogram(name, 50).With(label, value)
	n50 := fmt.Sprintf("%s_50", name)
	n90 := fmt.Sprintf("%s_90", name)
	n95 := fmt.Sprintf("%s_95", name)
	n99 := fmt.Sprintf("%s_99", name)
	quantiles := func() (p50, p90, p95, p99 float64) {
		err := cw.Send()
		if err != nil {
			t.Fatal(err)
		}
		svc.mtx.RLock()
		defer svc.mtx.RUnlock()
		p50 = svc.valuesReceived[n50]
		p90 = svc.valuesReceived[n90]
		p95 = svc.valuesReceived[n95]
		p99 = svc.valuesReceived[n99]
		return
	}
	if err := teststat.TestHistogram(histogram, quantiles, 0.01); err != nil {
		t.Fatal(err)
	}
	if err := testDimensions(svc, n50, label, value); err != nil {
		t.Fatal(err)
	}
	if err := testDimensions(svc, n90, label, value); err != nil {
		t.Fatal(err)
	}
	if err := testDimensions(svc, n95, label, value); err != nil {
		t.Fatal(err)
	}
	if err := testDimensions(svc, n99, label, value); err != nil {
		t.Fatal(err)
	}
}
