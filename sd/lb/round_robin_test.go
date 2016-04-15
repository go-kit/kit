package lb

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
)

func TestRoundRobin(t *testing.T) {
	var (
		method    = "hello"
		counts    = []int{0, 0, 0}
		endpoints = []endpoint.Endpoint{
			func(context.Context, interface{}) (interface{}, error) { counts[0]++; return struct{}{}, nil },
			func(context.Context, interface{}) (interface{}, error) { counts[1]++; return struct{}{}, nil },
			func(context.Context, interface{}) (interface{}, error) { counts[2]++; return struct{}{}, nil },
		}
		services = []sd.Service{
			sd.StaticService{method: endpoints[0]},
			sd.StaticService{method: endpoints[1]},
			sd.StaticService{method: endpoints[2]},
		}
	)

	subscriber := sd.StaticSubscriber(services)
	balancer := NewRoundRobin(subscriber)

	for i, want := range [][]int{
		{1, 0, 0},
		{1, 1, 0},
		{1, 1, 1},
		{2, 1, 1},
		{2, 2, 1},
		{2, 2, 2},
		{3, 2, 2},
	} {
		service, err := balancer.Service()
		if err != nil {
			t.Fatal(err)
		}
		endpoint, err := service.Endpoint(method)
		if err != nil {
			t.Fatal(err)
		}
		endpoint(context.Background(), struct{}{})
		if have := counts; !reflect.DeepEqual(want, have) {
			t.Fatalf("%d: want %v, have %v", i, want, have)
		}
	}
}

func TestRoundRobinNoEndpoints(t *testing.T) {
	subscriber := sd.StaticSubscriber{}
	balancer := NewRoundRobin(subscriber)
	_, err := balancer.Service()
	if want, have := ErrNoServices, err; want != have {
		t.Errorf("want %v, have %v", want, have)
	}
}
