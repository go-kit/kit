package lb

import (
	"math"
	"testing"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/service"
	"golang.org/x/net/context"
)

func TestRandom(t *testing.T) {
	var (
		n          = 7
		method     = "hello"
		endpoints  = make([]endpoint.Endpoint, n)
		services   = make([]service.Service, n)
		counts     = make([]int, n)
		seed       = int64(12345)
		iterations = 1000000
		want       = iterations / n
		tolerance  = want / 100 // 1%
	)

	for i := 0; i < n; i++ {
		i0 := i
		endpoints[i] = func(context.Context, interface{}) (interface{}, error) { counts[i0]++; return struct{}{}, nil }
		services[i] = service.Fixed{method: endpoints[i0]}
	}

	subscriber := sd.StaticSubscriber(services)
	balancer := NewRandom(subscriber, seed)

	for i := 0; i < iterations; i++ {
		service, _ := balancer.Service()
		endpoint, _ := service.Endpoint(method)
		endpoint(context.Background(), struct{}{})
	}

	for i, have := range counts {
		delta := int(math.Abs(float64(want - have)))
		if delta > tolerance {
			t.Errorf("%d: want %d, have %d, delta %d > %d tolerance", i, want, have, delta, tolerance)
		}
	}
}

func TestRandomNoEndpoints(t *testing.T) {
	subscriber := sd.StaticSubscriber{}
	balancer := NewRandom(subscriber, 1415926)
	_, err := balancer.Service()
	if want, have := ErrNoServices, err; want != have {
		t.Errorf("want %v, have %v", want, have)
	}

}
