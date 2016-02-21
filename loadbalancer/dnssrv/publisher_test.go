package dnssrv

import (
	"errors"
	"io"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

func TestPublisher(t *testing.T) {
	var (
		name    = "foo"
		ttl     = time.Second
		e       = func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil }
		factory = func(string) (endpoint.Endpoint, io.Closer, error) { return e, nil, nil }
		logger  = log.NewNopLogger()
	)

	p := NewPublisher(name, ttl, factory, logger)
	defer p.Stop()

	if _, err := p.Endpoints(); err != nil {
		t.Fatal(err)
	}
}

func TestBadLookup(t *testing.T) {
	var (
		name      = "some-name"
		ticker    = time.NewTicker(time.Second)
		lookups   = uint32(0)
		lookupSRV = func(string, string, string) (string, []*net.SRV, error) {
			atomic.AddUint32(&lookups, 1)
			return "", nil, errors.New("kaboom")
		}
		e       = func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil }
		factory = func(string) (endpoint.Endpoint, io.Closer, error) { return e, nil, nil }
		logger  = log.NewNopLogger()
	)

	p := NewPublisherDetailed(name, ticker, lookupSRV, factory, logger)
	defer p.Stop()

	endpoints, err := p.Endpoints()
	if err != nil {
		t.Error(err)
	}
	if want, have := 0, len(endpoints); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
	if want, have := uint32(1), atomic.LoadUint32(&lookups); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestBadFactory(t *testing.T) {
	var (
		name      = "some-name"
		ticker    = time.NewTicker(time.Second)
		addr      = &net.SRV{Target: "foo", Port: 1234}
		addrs     = []*net.SRV{addr}
		lookupSRV = func(a, b, c string) (string, []*net.SRV, error) { return "", addrs, nil }
		creates   = uint32(0)
		factory   = func(s string) (endpoint.Endpoint, io.Closer, error) {
			atomic.AddUint32(&creates, 1)
			return nil, nil, errors.New("kaboom")
		}
		logger = log.NewNopLogger()
	)

	p := NewPublisherDetailed(name, ticker, lookupSRV, factory, logger)
	defer p.Stop()

	endpoints, err := p.Endpoints()
	if err != nil {
		t.Error(err)
	}
	if want, have := 0, len(endpoints); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
	if want, have := uint32(1), atomic.LoadUint32(&creates); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestRefreshWithChange(t *testing.T) {
	t.Skip("TODO")
}

func TestRefreshNoChange(t *testing.T) {
	var (
		addr      = &net.SRV{Target: "my-target", Port: 5678}
		addrs     = []*net.SRV{addr}
		name      = "my-name"
		ticker    = time.NewTicker(time.Second)
		lookups   = uint32(0)
		lookupSRV = func(string, string, string) (string, []*net.SRV, error) {
			atomic.AddUint32(&lookups, 1)
			return "", addrs, nil
		}
		e       = func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil }
		factory = func(string) (endpoint.Endpoint, io.Closer, error) { return e, nil, nil }
		logger  = log.NewNopLogger()
	)

	ticker.Stop()
	tickc := make(chan time.Time)
	ticker.C = tickc

	p := NewPublisherDetailed(name, ticker, lookupSRV, factory, logger)
	defer p.Stop()

	if want, have := uint32(1), atomic.LoadUint32(&lookups); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	tickc <- time.Now()

	if want, have := uint32(2), atomic.LoadUint32(&lookups); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestRefreshResolveError(t *testing.T) {
	t.Skip("TODO")
}
