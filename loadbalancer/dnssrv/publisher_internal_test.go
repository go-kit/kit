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
		target = "my-target"
		port   = uint16(1234)
		addr   = &net.SRV{Target: target, Port: port}
		addrs  = []*net.SRV{addr}
		name   = "my-name"
		ttl    = time.Second
		logger = log.NewNopLogger()
		e      = func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil }
	)

	oldLookup := lookupSRV
	defer func() { lookupSRV = oldLookup }()
	lookupSRV = mockLookupSRV(addrs, nil, nil)

	factory := func(instance string) (endpoint.Endpoint, error) {
		if want, have := addr2instance(addr), instance; want != have {
			t.Errorf("want %q, have %q", want, have)
		}
		return e, nil
	}

	p, err := NewPublisher(name, ttl, factory, logger)
	if err != nil {
		t.Fatal(err)
	}
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
		factory = func(string) (endpoint.Endpoint, error) { return nil, errors.New("unreachable") }
		logger  = log.NewNopLogger()
	)

	if _, err := NewPublisher(name, ttl, factory, logger); err == nil {
		t.Fatal("wanted error, got none")
	}
}

func TestBadFactory(t *testing.T) {
	var (
		addr    = &net.SRV{Target: "foo", Port: 1234}
		addrs   = []*net.SRV{addr}
		name    = "some-name"
		ttl     = time.Second
		factory = func(string) (endpoint.Endpoint, error) { return nil, errors.New("kaboom") }
		logger  = log.NewNopLogger()
	)

	oldLookup := lookupSRV
	defer func() { lookupSRV = oldLookup }()
	lookupSRV = mockLookupSRV(addrs, nil, nil)

	p, err := NewPublisher(name, ttl, factory, logger)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Stop()

	endpoints, err := p.Endpoints()
	if err != nil {
		t.Fatal(err)
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
		factory = func(string) (endpoint.Endpoint, error) { return nil, errors.New("kaboom") }
		logger  = log.NewNopLogger()
	)

	oldTicker := newTicker
	defer func() { newTicker = oldTicker }()
	newTicker = func(time.Duration) *time.Ticker { return &time.Ticker{C: tick} }

	var resolves uint64
	oldLookup := lookupSRV
	defer func() { lookupSRV = oldLookup }()
	lookupSRV = mockLookupSRV(addrs, nil, &resolves)

	p, err := NewPublisher(name, ttl, factory, logger)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Stop()

	tick <- time.Now()
	if want, have := uint64(2), resolves; want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestRefreshResolveError(t *testing.T) {
	t.Skip("TODO")
}

func TestErrPublisherStopped(t *testing.T) {
	var (
		name    = "my-name"
		ttl     = time.Second
		factory = func(string) (endpoint.Endpoint, error) { return nil, errors.New("kaboom") }
		logger  = log.NewNopLogger()
	)

	oldLookup := lookupSRV
	defer func() { lookupSRV = oldLookup }()
	lookupSRV = mockLookupSRV([]*net.SRV{}, nil, nil)

	p, err := NewPublisher(name, ttl, factory, logger)
	if err != nil {
		t.Fatal(err)
	}

	p.Stop()
	_, have := p.Endpoints()
	if want := loadbalancer.ErrPublisherStopped; want != have {
		t.Fatalf("want %v, have %v", want, have)
	}
}

func mockLookupSRV(addrs []*net.SRV, err error, count *uint64) func(service, proto, name string) (string, []*net.SRV, error) {
	return func(service, proto, name string) (string, []*net.SRV, error) {
		if count != nil {
			atomic.AddUint64(count, 1)
		}
		return "", addrs, err
	}
}
