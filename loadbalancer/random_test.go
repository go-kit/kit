package loadbalancer_test

import (
	"math"
	"testing"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer"
	"github.com/go-kit/kit/loadbalancer/fixed"
)

func TestRandomDistribution(t *testing.T) {
	var (
		n          = 3
		endpoints  = make([]endpoint.Endpoint, n)
		counts     = make([]int, n)
		seed       = int64(123)
		ctx        = context.Background()
		iterations = 100000
		want       = iterations / n
		tolerance  = want / 100 // 1%
	)

	for i := 0; i < n; i++ {
		i0 := i
		endpoints[i] = func(context.Context, interface{}) (interface{}, error) { counts[i0]++; return struct{}{}, nil }
	}

	lb := loadbalancer.NewRandom(fixed.NewPublisher(endpoints), seed)

	for i := 0; i < iterations; i++ {
		e, err := lb.Endpoint()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := e(ctx, struct{}{}); err != nil {
			t.Error(err)
		}
	}

	for i, have := range counts {
		if math.Abs(float64(want-have)) > float64(tolerance) {
			t.Errorf("%d: want %d, have %d", i, want, have)
		}
	}
}

func TestRandomBadPublisher(t *testing.T) {
	t.Skip("TODO")
}

func TestRandomNoEndpoints(t *testing.T) {
	lb := loadbalancer.NewRandom(fixed.NewPublisher([]endpoint.Endpoint{}), 123)
	_, have := lb.Endpoint()
	if want := loadbalancer.ErrNoEndpoints; want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}
