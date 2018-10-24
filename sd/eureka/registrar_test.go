package eureka

import (
	"testing"
	"time"
)

func TestRegistrar(t *testing.T) {
	connection := &testConnection{
		errHeartbeat: errTest,
	}

	registrar1 := NewRegistrar(connection, instanceTest1, loggerTest)
	registrar2 := NewRegistrar(connection, instanceTest2, loggerTest)

	// Not registered.
	err := registrar1.Deregister()
	if err != nil {
		t.Errorf("error when deregistering %s", err)
	}
	if want, have := 0, len(connection.instances); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	// Register.
	err = registrar1.Register()
	if err != nil {
		t.Errorf("error when registering %s", err)
	}
	if want, have := 1, len(connection.instances); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	err = registrar2.Register()
	if err != nil {
		t.Errorf("error when registering %s", err)
	}
	if want, have := 2, len(connection.instances); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	// Deregister.
	err = registrar1.Deregister()
	if err != nil {
		t.Errorf("error when deregistering %s", err)
	}
	if want, have := 1, len(connection.instances); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	// Already registered.
	err = registrar1.Register()
	if err != nil {
		t.Errorf("error when registering %s", err)
	}
	if want, have := 2, len(connection.instances); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
	err = registrar1.Register()
	if err != nil {
		t.Errorf("error when registering %s", err)
	}
	if want, have := 2, len(connection.instances); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	// Wait for a heartbeat failure.
	time.Sleep(1010 * time.Millisecond)
	if want, have := 2, len(connection.instances); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
	err = registrar1.Deregister()
	if err != nil {
		t.Errorf("error when deregistering %s", err)
	}
	if want, have := 1, len(connection.instances); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestBadRegister(t *testing.T) {
	connection := &testConnection{
		errRegister: errTest,
	}

	registrar := NewRegistrar(connection, instanceTest1, loggerTest)
	registrar.Register()
	if want, have := 0, len(connection.instances); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestBadDeregister(t *testing.T) {
	connection := &testConnection{
		errDeregister: errTest,
	}

	registrar := NewRegistrar(connection, instanceTest1, loggerTest)
	err := registrar.Register()
	if err != nil {
		t.Errorf("error when registering %s", err)
	}
	if want, have := 1, len(connection.instances); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
	registrar.Deregister()
	if want, have := 1, len(connection.instances); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestExpiredInstance(t *testing.T) {
	connection := &testConnection{
		errHeartbeat: errNotFound,
	}

	registrar := NewRegistrar(connection, instanceTest1, loggerTest)
	err := registrar.Register()
	if err != nil {
		t.Errorf("error when registering %s", err)
	}

	// Wait for a heartbeat failure.
	time.Sleep(1010 * time.Millisecond)

	if want, have := 1, len(connection.instances); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}
