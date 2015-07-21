package loadbalancer

import (
	"fmt"
	"net"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
)

func TestDNSSRVPublisher(t *testing.T) {
	// Reset the vars when we're done
	oldResolve := resolve
	defer func() { resolve = oldResolve }()
	oldNewTicker := newTicker
	defer func() { newTicker = oldNewTicker }()

	// Set up a fixture and swap the vars
	a := []*net.SRV{
		{Target: "foo", Port: 123},
		{Target: "bar", Port: 456},
		{Target: "baz", Port: 789},
	}
	ticker := make(chan time.Time)
	resolve = func(string) ([]*net.SRV, string, error) { return a, fmt.Sprint(len(a)), nil }
	newTicker = func(time.Duration) *time.Ticker { return &time.Ticker{C: ticker} }

	// Construct endpoint
	m := map[string]int{}
	e := func(hostport string) endpoint.Endpoint {
		return func(context.Context, interface{}) (interface{}, error) {
			m[hostport]++
			return struct{}{}, nil
		}
	}

	// Build the publisher
	var (
		name         = "irrelevant"
		ttl          = time.Second
		makeEndpoint = func(hostport string) endpoint.Endpoint { return e(hostport) }
	)
	p := NewDNSSRVPublisher(name, ttl, makeEndpoint)
	defer p.Stop()

	// Subscribe
	c := make(chan []endpoint.Endpoint, 1)
	p.Subscribe(c)
	defer p.Unsubscribe(c)

	// Invoke all of the endpoints
	for _, e := range <-c {
		e(context.Background(), struct{}{})
	}

	// Make sure we invoked what we expected to
	for _, addr := range a {
		hostport := addr2hostport(addr)
		if want, have := 1, m[hostport]; want != have {
			t.Errorf("%q: want %d, have %d", name, want, have)
		}
		delete(m, hostport)
	}
	if want, have := 0, len(m); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	// Reset the fixture, trigger the timer, count the endpoints
	a = []*net.SRV{}
	ticker <- time.Now()
	if want, have := len(a), len(<-c); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}
