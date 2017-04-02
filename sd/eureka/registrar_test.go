package eureka

import (
	"testing"
	"time"

	"github.com/hudl/fargo"
)

func TestRegistrar(t *testing.T) {
	client := &testClient{
		instances:    []*fargo.Instance{},
		errHeartbeat: errTest,
	}

	r := NewRegistrar(client, instanceTest1, loggerTest)
	if want, have := 0, len(client.instances); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	// Not registered.
	r.Deregister()
	if want, have := 0, len(client.instances); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	// Register.
	r.Register()
	if want, have := 1, len(client.instances); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	// Deregister.
	r.Deregister()
	if want, have := 0, len(client.instances); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	// Already registered.
	r.Register()
	if want, have := 1, len(client.instances); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
	r.Register()
	if want, have := 1, len(client.instances); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	// Wait for a heartbeat failure.
	time.Sleep(time.Second)
	if want, have := 1, len(client.instances); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
	r.Deregister()
	if want, have := 0, len(client.instances); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}
