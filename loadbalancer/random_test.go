package loadbalancer_test

import (
	"math"
	"testing"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer"
)

func TestRandomStrategy(t *testing.T) {
	var (
		counts    = []int{0, 0, 0}
		endpoints = []endpoint.Endpoint{
			func(context.Context, interface{}) (interface{}, error) { counts[0]++; return struct{}{}, nil },
			func(context.Context, interface{}) (interface{}, error) { counts[1]++; return struct{}{}, nil },
			func(context.Context, interface{}) (interface{}, error) { counts[2]++; return struct{}{}, nil },
		}
	)

	p := newMockPublisher([]endpoint.Endpoint{})
	s := loadbalancer.Random(p)
	defer s.Stop()

	_, have := s.Next()
	if want := loadbalancer.ErrNoEndpoints; want != have {
		t.Errorf("want %v, have %v", want, have)
	}

	p.replace([]endpoint.Endpoint{endpoints[0]})
	for i := 0; i < len(counts); i++ {
		e, _ := s.Next()
		e(context.Background(), struct{}{})
	}
	if want, have := len(counts), counts[0]; want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	counts[0] = 0
	p.replace(endpoints)
	n := 10000
	for i := 0; i < n; i++ {
		e, _ := s.Next()
		e(context.Background(), struct{}{})
	}
	want := float64(n) / float64(len(counts))
	tolerance := float64(n) / 100.0 // 1% error
	for _, count := range counts {
		if have := float64(count); math.Abs(want-have) > tolerance {
			t.Errorf("want %.0f, have %.0f", want, have)
		}
	}
}
