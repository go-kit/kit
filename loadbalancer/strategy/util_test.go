package strategy_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-kit/kit/loadbalancer"
)

func assertLoadBalancerNotEmpty(t *testing.T, lb loadbalancer.LoadBalancer) {
	if err := within(10*time.Millisecond, func() bool {
		return lb.Count() > 0
	}); err != nil {
		t.Fatal("Publisher never updated endpoints")
	}
}

func within(d time.Duration, f func() bool) error {
	var (
		deadline = time.After(d)
		ticker   = time.NewTicker(d / 10)
	)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if f() {
				return nil
			}
		case <-deadline:
			return fmt.Errorf("deadline exceeded")
		}
	}
}
