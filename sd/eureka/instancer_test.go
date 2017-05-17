package eureka

import (
	"io"
	"testing"
	"time"

	"github.com/hudl/fargo"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
)

var _ sd.Instancer = &Instancer{} // API check

func TestInstancer(t *testing.T) {
	factory := func(string) (endpoint.Endpoint, io.Closer, error) {
		return endpoint.Nop, nil, nil
	}

	connection := &testConnection{
		instances:      []*fargo.Instance{instanceTest1},
		application:    appUpdateTest,
		errApplication: nil,
	}

	instancer := NewInstancer(connection, appNameTest, loggerTest)
	defer instancer.Stop()
	endpointer := sd.NewEndpointer(instancer, factory, loggerTest)

	endpoints, err := endpointer.Endpoints()
	if err != nil {
		t.Fatal(err)
	}

	if want, have := 1, len(endpoints); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestInstancerScheduleUpdates(t *testing.T) {
	factory := func(string) (endpoint.Endpoint, io.Closer, error) {
		return endpoint.Nop, nil, nil
	}

	connection := &testConnection{
		instances:      []*fargo.Instance{instanceTest1},
		application:    appUpdateTest,
		errApplication: nil,
	}

	instancer := NewInstancer(connection, appNameTest, loggerTest)
	defer instancer.Stop()
	endpointer := sd.NewEndpointer(instancer, factory, loggerTest)

	endpoints, _ := endpointer.Endpoints()
	if want, have := 1, len(endpoints); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	time.Sleep(50 * time.Millisecond)

	endpoints, _ = endpointer.Endpoints()
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

	instancer := NewInstancer(connection, appNameTest, loggerTest)
	defer instancer.Stop()
	endpointer := sd.NewEndpointer(instancer, factory, loggerTest)

	endpoints, err := endpointer.Endpoints()
	if err != nil {
		t.Fatal(err)
	}

	if want, have := 0, len(endpoints); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestBadInstancerInstances(t *testing.T) {
	factory := func(string) (endpoint.Endpoint, io.Closer, error) {
		return endpoint.Nop, nil, nil
	}

	connection := &testConnection{
		instances:      []*fargo.Instance{},
		errInstances:   errTest,
		application:    appUpdateTest,
		errApplication: nil,
	}

	instancer := NewInstancer(connection, appNameTest, loggerTest)
	defer instancer.Stop()
	endpointer := sd.NewEndpointer(instancer, factory, loggerTest)

	endpoints, err := endpointer.Endpoints()
	if err != nil {
		t.Fatal(err)
	}

	if want, have := 0, len(endpoints); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestBadInstancerScheduleUpdates(t *testing.T) {
	factory := func(string) (endpoint.Endpoint, io.Closer, error) {
		return endpoint.Nop, nil, nil
	}

	connection := &testConnection{
		instances:      []*fargo.Instance{instanceTest1},
		application:    appUpdateTest,
		errApplication: errTest,
	}

	instancer := NewInstancer(connection, appNameTest, loggerTest)
	defer instancer.Stop()
	endpointer := sd.NewEndpointer(instancer, factory, loggerTest)

	endpoints, err := endpointer.Endpoints()
	if err != nil {
		t.Error(err)
	}
	if want, have := 1, len(endpoints); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	time.Sleep(50 * time.Millisecond)

	endpoints, err = endpointer.Endpoints()
	if err != nil {
		t.Error(err)
	}
	if want, have := 1, len(endpoints); want != have {
		t.Errorf("want %v, have %v", want, have)
	}
}
