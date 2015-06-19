package loadbalancer_test

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer"
)

func TestRoundRobin(t *testing.T) {
	var (
		counts    = []int{0, 0, 0}
		endpoints = []endpoint.Endpoint{
			func(context.Context, interface{}) (interface{}, error) { counts[0]++; return struct{}{}, nil },
			func(context.Context, interface{}) (interface{}, error) { counts[1]++; return struct{}{}, nil },
			func(context.Context, interface{}) (interface{}, error) { counts[2]++; return struct{}{}, nil },
		}
	)

	p := newMockPublisher([]endpoint.Endpoint{})
	s := loadbalancer.RoundRobin(p)
	defer s.Stop()

	_, have := s.Next()
	if want := loadbalancer.ErrNoEndpoints; want != have {
		t.Errorf("want %v, have %v", want, have)
	}

	p.replace(endpoints)
	for i, want := range [][]int{
		{1, 0, 0},
		{1, 1, 0},
		{1, 1, 1},
		{2, 1, 1},
		{2, 2, 1},
		{2, 2, 2},
		{3, 2, 2},
	} {
		e, _ := s.Next()
		e(context.Background(), struct{}{})
		if have := counts; !reflect.DeepEqual(want, have) {
			t.Errorf("%d: want %v, have %v", i+1, want, have)
		}
	}
}
