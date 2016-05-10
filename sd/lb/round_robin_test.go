package lb

import (
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/service"
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
		services = []service.Service{
			service.Fixed{method: endpoints[0]},
			service.Fixed{method: endpoints[1]},
			service.Fixed{method: endpoints[2]},
		}
	)

	subscriber := sd.FixedSubscriber(services)
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
	subscriber := sd.FixedSubscriber{}
	balancer := NewRoundRobin(subscriber)
	_, err := balancer.Service()
	if want, have := ErrNoServices, err; want != have {
		t.Errorf("want %v, have %v", want, have)
	}
}

func TestRoundRobinNoRace(t *testing.T) {
	balancer := NewRoundRobin(sd.FixedSubscriber([]service.Service{
		service.Fixed{"a": endpoint.Nop},
		service.Fixed{"b": endpoint.Nop},
		service.Fixed{"c": endpoint.Nop},
	}))

	var (
		n     = 100
		done  = make(chan struct{})
		wg    sync.WaitGroup
		count uint64
	)

	wg.Add(n)

	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					_, _ = balancer.Service()
					atomic.AddUint64(&count, 1)
				}
			}
		}()
	}

	time.Sleep(time.Second)
	close(done)
	wg.Wait()

	t.Logf("made %d calls", atomic.LoadUint64(&count))
}
