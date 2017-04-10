package eureka

import (
	"io"
	"testing"
	"time"

	"github.com/hudl/fargo"

	"github.com/go-kit/kit/endpoint"
)

func TestSubscriber(t *testing.T) {
	factory := func(string) (endpoint.Endpoint, io.Closer, error) {
		return endpoint.Nop, nil, nil
	}

	connection := &testConnection{
		instances:      []*fargo.Instance{instanceTest1},
		application:    appUpdateTest,
		errApplication: nil,
	}

	subscriber := NewSubscriber(connection, appNameTest, factory, loggerTest)
	defer subscriber.Stop()

	endpoints, err := subscriber.Endpoints()
	if err != nil {
		t.Fatal(err)
	}

	if want, have := 1, len(endpoints); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestSubscriberScheduleUpdates(t *testing.T) {
	factory := func(string) (endpoint.Endpoint, io.Closer, error) {
		return endpoint.Nop, nil, nil
	}

	connection := &testConnection{
		instances:      []*fargo.Instance{instanceTest1},
		application:    appUpdateTest,
		errApplication: nil,
	}

	subscriber := NewSubscriber(connection, appNameTest, factory, loggerTest)
	defer subscriber.Stop()

	endpoints, _ := subscriber.Endpoints()
	if want, have := 1, len(endpoints); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	time.Sleep(50 * time.Millisecond)

	endpoints, _ = subscriber.Endpoints()
	if want, have := 2, len(endpoints); want != have {
		t.Errorf("want %v, have %v", want, have)
	}
}

func TestBadFactory(t *testing.T) {
	factory := func(string) (endpoint.Endpoint, io.Closer, error) {
		return nil, nil, errTest
	}

	connection := &testConnection{
		instances:      []*fargo.Instance{instanceTest1},
		application:    appUpdateTest,
		errApplication: nil,
	}

	subscriber := NewSubscriber(connection, appNameTest, factory, loggerTest)
	defer subscriber.Stop()

	endpoints, err := subscriber.Endpoints()
	if err != nil {
		t.Fatal(err)
	}

	if want, have := 0, len(endpoints); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestBadSubscriberInstances(t *testing.T) {
	factory := func(string) (endpoint.Endpoint, io.Closer, error) {
		return endpoint.Nop, nil, nil
	}

	connection := &testConnection{
		instances:      []*fargo.Instance{},
		errInstances:   errTest,
		application:    appUpdateTest,
		errApplication: nil,
	}

	subscriber := NewSubscriber(connection, appNameTest, factory, loggerTest)
	defer subscriber.Stop()

	endpoints, err := subscriber.Endpoints()
	if err != nil {
		t.Fatal(err)
	}

	if want, have := 0, len(endpoints); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestBadSubscriberScheduleUpdates(t *testing.T) {
	factory := func(string) (endpoint.Endpoint, io.Closer, error) {
		return endpoint.Nop, nil, nil
	}

	connection := &testConnection{
		instances:      []*fargo.Instance{instanceTest1},
		application:    appUpdateTest,
		errApplication: errTest,
	}

	subscriber := NewSubscriber(connection, appNameTest, factory, loggerTest)
	defer subscriber.Stop()

	endpoints, err := subscriber.Endpoints()
	if err != nil {
		t.Error(err)
	}
	if want, have := 1, len(endpoints); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	time.Sleep(50 * time.Millisecond)

	endpoints, err = subscriber.Endpoints()
	if err != nil {
		t.Error(err)
	}
	if want, have := 1, len(endpoints); want != have {
		t.Errorf("want %v, have %v", want, have)
	}
}
