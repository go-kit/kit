package dnssrv

import (
	"errors"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer"
	"github.com/go-kit/kit/log"
)

func TestPublisher(t *testing.T) {
	var (
		name    = "foo"
		ttl     = time.Second
		e       = func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil }
		c       = make(chan struct{})
		factory = func(string) (endpoint.Endpoint, loadbalancer.Closer, error) { return e, c, nil }
		logger  = log.NewNopLogger()
	)

	p := NewPublisher(name, ttl, factory, logger)
	defer p.Stop()

	if _, err := p.Endpoints(); err != nil {
		t.Fatal(err)
	}
}

func TestBadLookup(t *testing.T) {
	oldLookup := lookupSRV
	defer func() { lookupSRV = oldLookup }()
	lookupSRV = mockLookupSRV([]*net.SRV{}, errors.New("kaboom"), nil)

	var (
		name    = "some-name"
		ttl     = time.Second
		e       = func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil }
		c       = make(chan struct{})
		factory = func(string) (endpoint.Endpoint, loadbalancer.Closer, error) { return e, c, nil }
		logger  = log.NewNopLogger()
	)

	p := NewPublisher(name, ttl, factory, logger)
	defer p.Stop()

	endpoints, err := p.Endpoints()
	if err != nil {
		t.Error(err)
	}
	if want, have := 0, len(endpoints); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestBadFactory(t *testing.T) {
	var (
		addr    = &net.SRV{Target: "foo", Port: 1234}
		addrs   = []*net.SRV{addr}
		name    = "some-name"
		ttl     = time.Second
		factory = func(string) (endpoint.Endpoint, loadbalancer.Closer, error) { return nil, nil, errors.New("kaboom") }
		logger  = log.NewNopLogger()
	)

	oldLookup := lookupSRV
	defer func() { lookupSRV = oldLookup }()
	lookupSRV = mockLookupSRV(addrs, nil, nil)

	p := NewPublisher(name, ttl, factory, logger)
	defer p.Stop()

	endpoints, err := p.Endpoints()
	if err != nil {
		t.Error(err)
	}
	if want, have := 0, len(endpoints); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}

func TestRefreshWithChange(t *testing.T) {
	t.Skip("TODO")
}

func TestRefreshNoChange(t *testing.T) {
	var (
		tick    = make(chan time.Time)
		target  = "my-target"
		port    = uint16(5678)
		addr    = &net.SRV{Target: target, Port: port}
		addrs   = []*net.SRV{addr}
		name    = "my-name"
		ttl     = time.Second
		factory = func(string) (endpoint.Endpoint, loadbalancer.Closer, error) { return nil, nil, errors.New("kaboom") }
		logger  = log.NewNopLogger()
	)

	oldTicker := newTicker
	defer func() { newTicker = oldTicker }()
	newTicker = func(time.Duration) *time.Ticker { return &time.Ticker{C: tick} }

	var resolves uint64
	oldLookup := lookupSRV
	defer func() { lookupSRV = oldLookup }()
	lookupSRV = mockLookupSRV(addrs, nil, &resolves)

	p := NewPublisher(name, ttl, factory, logger)
	defer p.Stop()

	tick <- time.Now()
	if want, have := uint64(2), resolves; want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestRefreshResolveError(t *testing.T) {
	t.Skip("TODO")
}

func mockLookupSRV(addrs []*net.SRV, err error, count *uint64) func(service, proto, name string) (string, []*net.SRV, error) {
	return func(service, proto, name string) (string, []*net.SRV, error) {
		if count != nil {
			atomic.AddUint64(count, 1)
		}
		return "", addrs, err
	}
}
