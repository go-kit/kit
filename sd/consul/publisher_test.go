package consul

import (
	"testing"

	stdconsul "github.com/hashicorp/consul/api"

	"github.com/go-kit/kit/log"
)

func TestPublisher(t *testing.T) {
	client := newTestClient([]*stdconsul.ServiceEntry{})
	p := NewPublisher(client, testRegistration, log.NewNopLogger())
	if want, have := 0, len(client.entries); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	p.Publish()
	if want, have := 1, len(client.entries); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	p.Unpublish()
	if want, have := 0, len(client.entries); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}
