package loadbalancer

import (
	"testing"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
)

func TestEndpointCache(t *testing.T) {
	endpoints := []endpoint.Endpoint{
		func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil },
		func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil },
		func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil },
	}

	p := NewStaticPublisher(endpoints)
	defer p.Stop()

	c := newEndpointCache(p)
	defer c.stop()

	if want, have := len(endpoints), len(c.get()); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}
