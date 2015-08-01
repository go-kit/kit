package strategy_test

import (
	"math"
	"testing"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer/publisher/static"
	"github.com/go-kit/kit/loadbalancer/strategy"
	"golang.org/x/net/context"
)

func TestRandom(t *testing.T) {
	p := static.NewPublisher([]endpoint.Endpoint{})
	defer p.Stop()

	lb := strategy.Random(p)
	if _, err := lb.Get(); err == nil {
		t.Error("want error, got none")
	}

	counts := []int{0, 0, 0}
	p.Replace([]endpoint.Endpoint{
		func(context.Context, interface{}) (interface{}, error) { counts[0]++; return struct{}{}, nil },
		func(context.Context, interface{}) (interface{}, error) { counts[1]++; return struct{}{}, nil },
		func(context.Context, interface{}) (interface{}, error) { counts[2]++; return struct{}{}, nil },
	})
	assertLoadBalancerNotEmpty(t, lb)

	n := 10000
	for i := 0; i < n; i++ {
		e, _ := lb.Get()
		e(context.Background(), struct{}{})
	}

	want := float64(n) / float64(len(counts))
	tolerance := (want / 100.0) * 5 // 5%
	for _, have := range counts {
		if math.Abs(want-float64(have)) > tolerance {
			t.Errorf("want %.0f, have %d", want, have)
		}
	}
}
