package consul

import (
	"bytes"
	"testing"

	stdconsul "github.com/hashicorp/consul/api"

	"github.com/go-kit/kit/log"
)

func TestPublisher(t *testing.T) {
	client := newTestClient([]*stdconsul.ServiceEntry{})
	var buf bytes.Buffer
	p := NewPublisher(client, testRegistration, log.NewLogfmtLogger(&buf))

	p.Publish()
	if want, have := 0, buf.Len(); want != have {
		t.Error(buf.String())
	}

	buf.Reset()

	p.Unpublish()
	if want, have := 0, buf.Len(); want != have {
		t.Error(buf.String())
	}
}
