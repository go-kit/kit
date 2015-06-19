package loadbalancer_test

import (
	"testing"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer"
)

func TestStaticPublisher(t *testing.T) {
	for _, n := range []int{0, 1, 3} {
		endpoints := []endpoint.Endpoint{}
		for i := 0; i < n; i++ {
			endpoints = append(endpoints, func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil })
		}

		c := make(chan []endpoint.Endpoint, 1)
		p := loadbalancer.NewStaticPublisher(endpoints)
		p.Subscribe(c)
		if want, have := n, len(<-c); want != have {
			t.Errorf("want %d, have %d", want, have)
		}
		p.Unsubscribe(c)
	}
}
