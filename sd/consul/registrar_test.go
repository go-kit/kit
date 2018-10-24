package consul

import (
	"testing"

	"github.com/go-kit/kit/log"
	stdconsul "github.com/hashicorp/consul/api"
)

func TestRegistrar(t *testing.T) {
	client := newTestClient([]*stdconsul.ServiceEntry{})
	p := NewRegistrar(client, testRegistration, log.NewNopLogger())
	if want, have := 0, len(client.entries); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	err := p.Register()
	if err != nil {
		t.Errorf("error when registering %s", err)
	}
	if want, have := 1, len(client.entries); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	err = p.Deregister()
	if err != nil {
		t.Errorf("error when deregistering %s", err)
	}
	if want, have := 0, len(client.entries); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}
