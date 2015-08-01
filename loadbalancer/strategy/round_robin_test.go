package strategy_test

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer/publisher/static"
	"github.com/go-kit/kit/loadbalancer/strategy"
)

func TestRoundRobin(t *testing.T) {
	p := static.NewPublisher([]endpoint.Endpoint{})
	defer p.Stop()

	lb := strategy.RoundRobin(p)
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

	for i, want := range [][]int{
		{1, 0, 0},
		{1, 1, 0},
		{1, 1, 1},
		{2, 1, 1},
		{2, 2, 1},
		{2, 2, 2},
		{3, 2, 2},
	} {
		e, _ := lb.Get()
		e(context.Background(), struct{}{})
		if have := counts; !reflect.DeepEqual(want, have) {
			t.Errorf("%d: want %v, have %v", i+1, want, have)
		}
	}
}
