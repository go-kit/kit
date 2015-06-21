package loadbalancer_test

import (
	"testing"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer"
)

func TestStaticPublisher(t *testing.T) {
	endpoints := []endpoint.Endpoint{
		func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil },
	}
	p := loadbalancer.NewStaticPublisher(endpoints)
	defer p.Stop()

	c := make(chan []endpoint.Endpoint, 1)
	p.Subscribe(c)
	if want, have := len(endpoints), len(<-c); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	endpoints = []endpoint.Endpoint{}
	p.Replace(endpoints)
	if want, have := len(endpoints), len(<-c); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}
