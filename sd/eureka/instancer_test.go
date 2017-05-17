package eureka

import (
	"testing"
	"time"

	"github.com/hudl/fargo"

	"github.com/go-kit/kit/sd"
)

var _ sd.Instancer = &Instancer{} // API check

func TestInstancer(t *testing.T) {
	connection := &testConnection{
		instances:      []*fargo.Instance{instanceTest1},
		application:    appUpdateTest,
		errApplication: nil,
	}

	instancer := NewInstancer(connection, appNameTest, loggerTest)
	defer instancer.Stop()

	state := instancer.State()
	if state.Err != nil {
		t.Fatal(state.Err)
	}

	if want, have := 1, len(state.Instances); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestInstancerScheduleUpdates(t *testing.T) {
	connection := &testConnection{
		instances:      []*fargo.Instance{instanceTest1},
		application:    appUpdateTest,
		errApplication: nil,
	}

	instancer := NewInstancer(connection, appNameTest, loggerTest)
	defer instancer.Stop()

	state := instancer.State()
	if want, have := 1, len(state.Instances); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	time.Sleep(50 * time.Millisecond)

	state = instancer.State()
	if want, have := 2, len(state.Instances); want != have {
		t.Errorf("want %v, have %v", want, have)
	}
}

func TestBadInstancerInstances(t *testing.T) {
	connection := &testConnection{
		instances:      []*fargo.Instance{},
		errInstances:   errTest,
		application:    appUpdateTest,
		errApplication: nil,
	}

	instancer := NewInstancer(connection, appNameTest, loggerTest)
	defer instancer.Stop()

	state := instancer.State()
	if state.Err == nil {
		t.Fatal("expecting error")
	}

	if want, have := 0, len(state.Instances); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestBadInstancerScheduleUpdates(t *testing.T) {
	connection := &testConnection{
		instances:      []*fargo.Instance{instanceTest1},
		application:    appUpdateTest,
		errApplication: errTest,
	}

	instancer := NewInstancer(connection, appNameTest, loggerTest)
	defer instancer.Stop()

	state := instancer.State()
	if state.Err != nil {
		t.Error(state.Err)
	}
	if want, have := 1, len(state.Instances); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	time.Sleep(50 * time.Millisecond)

	state = instancer.State()
	if state.Err == nil {
		t.Fatal("expecting error")
	}

	if want, have := 0, len(state.Instances); want != have {
		t.Errorf("want %v, have %v", want, have)
	}
}
