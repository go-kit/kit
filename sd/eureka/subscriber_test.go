package eureka

import (
	"io"
	"testing"
	"time"

	"github.com/go-kit/kit/endpoint"
	stdeureka "github.com/hudl/fargo"
)

func TestSubscriber(t *testing.T) {
	factory := func(string) (endpoint.Endpoint, io.Closer, error) {
		return endpoint.Nop, nil, nil
	}

	client := &testClient{
		instances:      []*stdeureka.Instance{instanceTest1},
		application:    applicationTest,
		errApplication: nil,
	}

	s := NewSubscriber(client, factory, loggerTest, instanceTest1.App)
	defer s.Stop()

	endpoints, err := s.Endpoints()
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

	client := &testClient{
		instances:      []*stdeureka.Instance{instanceTest1},
		application:    applicationTest,
		errApplication: nil,
	}

	s := NewSubscriber(client, factory, loggerTest, instanceTest1.App)
	defer s.Stop()

	endpoints, _ := s.Endpoints()
	if want, have := 1, len(endpoints); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	time.Sleep(50 * time.Millisecond)

	endpoints, _ = s.Endpoints()
	if want, have := 2, len(endpoints); want != have {
		t.Errorf("want %v, have %v", want, have)
	}
}

func TestBadFactory(t *testing.T) {
	factory := func(string) (endpoint.Endpoint, io.Closer, error) {
		return nil, nil, errTest
	}

	client := &testClient{
		instances: []*stdeureka.Instance{instanceTest1},
	}

	s := NewSubscriber(client, factory, loggerTest, instanceTest1.App)
	defer s.Stop()

	endpoints, err := s.Endpoints()
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

	client := &testClient{
		errInstances:   errTest,
		application:    applicationTest,
		errApplication: nil,
	}

	s := NewSubscriber(client, factory, loggerTest, instanceTest1.App)
	defer s.Stop()

	endpoints, err := s.Endpoints()
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

	client := &testClient{
		instances:      []*stdeureka.Instance{instanceTest1},
		application:    applicationTest,
		errApplication: errTest,
	}

	s := NewSubscriber(client, factory, loggerTest, instanceTest1.App)
	defer s.Stop()

	endpoints, err := s.Endpoints()
	if err != nil {
		t.Error(err)
	}
	if want, have := 1, len(endpoints); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	time.Sleep(50 * time.Millisecond)

	endpoints, err = s.Endpoints()
	if err != nil {
		t.Error(err)
	}
	if want, have := 1, len(endpoints); want != have {
		t.Errorf("want %v, have %v", want, have)
	}
}
