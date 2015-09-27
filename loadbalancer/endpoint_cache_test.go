package loadbalancer_test

import (
	"io"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer"
	"github.com/go-kit/kit/log"
)

func TestEndpointCache(t *testing.T) {
	var (
		e  = func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil }
		ca = make(closer)
		cb = make(closer)
		c  = map[string]io.Closer{"a": ca, "b": cb}
		f  = func(s string) (endpoint.Endpoint, io.Closer, error) { return e, c[s], nil }
		ec = loadbalancer.NewEndpointCache(f, log.NewNopLogger())
	)

	// Populate
	ec.Replace([]string{"a", "b"})
	select {
	case <-ca:
		t.Errorf("endpoint a closed, not good")
	case <-cb:
		t.Errorf("endpoint b closed, not good")
	case <-time.After(time.Millisecond):
		t.Logf("no closures yet, good")
	}

	// Duplicate, should be no-op
	ec.Replace([]string{"a", "b"})
	select {
	case <-ca:
		t.Errorf("endpoint a closed, not good")
	case <-cb:
		t.Errorf("endpoint b closed, not good")
	case <-time.After(time.Millisecond):
		t.Logf("no closures yet, good")
	}

	// Delete b
	go ec.Replace([]string{"a"})
	select {
	case <-ca:
		t.Errorf("endpoint a closed, not good")
	case <-cb:
		t.Logf("endpoint b closed, good")
	case <-time.After(time.Millisecond):
		t.Errorf("didn't close the deleted instance in time")
	}

	// Delete a
	go ec.Replace([]string{""})
	select {
	// case <-cb: will succeed, as it's closed
	case <-ca:
		t.Logf("endpoint a closed, good")
	case <-time.After(time.Millisecond):
		t.Errorf("didn't close the deleted instance in time")
	}
}

type closer chan struct{}

func (c closer) Close() error { close(c); return nil }
