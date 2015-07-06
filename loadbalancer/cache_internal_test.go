package loadbalancer

import (
	"runtime"
	"testing"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
)

func TestCache(t *testing.T) {
	e := func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil }
	endpoints := []endpoint.Endpoint{e}

	p := NewStaticPublisher(endpoints)
	defer p.Stop()

	c := newCache(p)
	defer c.stop()

	for _, n := range []int{2, 10, 0} {
		endpoints = make([]endpoint.Endpoint, n)
		for i := 0; i < n; i++ {
			endpoints[i] = e
		}
		p.Replace(endpoints)
		runtime.Gosched()
		if want, have := len(endpoints), len(c.get()); want != have {
			t.Errorf("want %d, have %d", want, have)
		}
	}
}
