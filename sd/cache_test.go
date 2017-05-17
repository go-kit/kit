package sd

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

func TestCache(t *testing.T) {
	var (
		ca    = make(closer)
		cb    = make(closer)
		c     = map[string]io.Closer{"a": ca, "b": cb}
		f     = func(instance string) (endpoint.Endpoint, io.Closer, error) { return endpoint.Nop, c[instance], nil }
		cache = newEndpointCache(f, log.NewNopLogger(), endpointerOptions{})
	)

	// Populate
	cache.Update(Event{Instances: []string{"a", "b"}})
	select {
	case <-ca:
		t.Errorf("endpoint a closed, not good")
	case <-cb:
		t.Errorf("endpoint b closed, not good")
	case <-time.After(time.Millisecond):
		t.Logf("no closures yet, good")
	}
	assertEndpointsLen(t, cache, 2)

	// Duplicate, should be no-op
	cache.Update(Event{Instances: []string{"a", "b"}})
	select {
	case <-ca:
		t.Errorf("endpoint a closed, not good")
	case <-cb:
		t.Errorf("endpoint b closed, not good")
	case <-time.After(time.Millisecond):
		t.Logf("no closures yet, good")
	}
	assertEndpointsLen(t, cache, 2)

	// Delete b
	go cache.Update(Event{Instances: []string{"a"}})
	select {
	case <-ca:
		t.Errorf("endpoint a closed, not good")
	case <-cb:
		t.Logf("endpoint b closed, good")
	case <-time.After(time.Second):
		t.Errorf("didn't close the deleted instance in time")
	}
	assertEndpointsLen(t, cache, 1)

	// Delete a
	go cache.Update(Event{Instances: []string{}})
	select {
	// case <-cb: will succeed, as it's closed
	case <-ca:
		t.Logf("endpoint a closed, good")
	case <-time.After(time.Second):
		t.Errorf("didn't close the deleted instance in time")
	}
	assertEndpointsLen(t, cache, 0)
}

func TestBadFactory(t *testing.T) {
	cache := newEndpointCache(func(string) (endpoint.Endpoint, io.Closer, error) {
		return nil, nil, errors.New("bad factory")
	}, log.NewNopLogger(), endpointerOptions{})

	cache.Update(Event{Instances: []string{"foo:1234", "bar:5678"}})
	assertEndpointsLen(t, cache, 0)
}

func assertEndpointsLen(t *testing.T, cache *endpointCache, l int) {
	endpoints, err := cache.Endpoints()
	if err != nil {
		t.Errorf("unexpected error %v", err)
		return
	}
	if want, have := l, len(endpoints); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

type closer chan struct{}

func (c closer) Close() error { close(c); return nil }
