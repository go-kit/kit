package loadbalancer_test

import (
	"reflect"
	"runtime"
	"testing"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer"
	"golang.org/x/net/context"
)

func TestRoundRobin(t *testing.T) {
	p := loadbalancer.NewStaticPublisher([]endpoint.Endpoint{})
	defer p.Stop()

	lb := loadbalancer.RoundRobin(p)
	if _, err := lb.Get(); err == nil {
		t.Error("want error, got none")
	}

	counts := []int{0, 0, 0}
	p.Replace([]endpoint.Endpoint{
		func(context.Context, interface{}) (interface{}, error) { counts[0]++; return struct{}{}, nil },
		func(context.Context, interface{}) (interface{}, error) { counts[1]++; return struct{}{}, nil },
		func(context.Context, interface{}) (interface{}, error) { counts[2]++; return struct{}{}, nil },
	})
	runtime.Gosched()

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
