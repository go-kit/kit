package lb_test

import (
	"errors"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
	loadbalancer "github.com/go-kit/kit/sd/lb"
)

func TestRetryMaxTotalFail(t *testing.T) {
	var (
		endpoints = sd.FixedSubscriber{} // no endpoints
		lb        = loadbalancer.NewRoundRobin(endpoints)
		retry     = loadbalancer.Retry(999, time.Second, lb) // lots of retries
		ctx       = context.Background()
	)
	if _, err := retry(ctx, struct{}{}); err == nil {
		t.Errorf("expected error, got none") // should fail
	}
}

func TestRetryMaxPartialFail(t *testing.T) {
	var (
		endpoints = []endpoint.Endpoint{
			func(context.Context, interface{}) (interface{}, error) { return nil, errors.New("error one") },
			func(context.Context, interface{}) (interface{}, error) { return nil, errors.New("error two") },
			func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil /* OK */ },
		}
		subscriber = sd.FixedSubscriber{
			0: endpoints[0],
			1: endpoints[1],
			2: endpoints[2],
		}
		retries = len(endpoints) - 1 // not quite enough retries
		lb      = loadbalancer.NewRoundRobin(subscriber)
		ctx     = context.Background()
	)
	if _, err := loadbalancer.Retry(retries, time.Second, lb)(ctx, struct{}{}); err == nil {
		t.Errorf("expected error, got none")
	}
}

func TestRetryMaxSuccess(t *testing.T) {
	var (
		endpoints = []endpoint.Endpoint{
			func(context.Context, interface{}) (interface{}, error) { return nil, errors.New("error one") },
			func(context.Context, interface{}) (interface{}, error) { return nil, errors.New("error two") },
			func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil /* OK */ },
		}
		subscriber = sd.FixedSubscriber{
			0: endpoints[0],
			1: endpoints[1],
			2: endpoints[2],
		}
		retries = len(endpoints) // exactly enough retries
		lb      = loadbalancer.NewRoundRobin(subscriber)
		ctx     = context.Background()
	)
	if _, err := loadbalancer.Retry(retries, time.Second, lb)(ctx, struct{}{}); err != nil {
		t.Error(err)
	}
}

func TestRetryTimeout(t *testing.T) {
	var (
		step    = make(chan struct{})
		e       = func(context.Context, interface{}) (interface{}, error) { <-step; return struct{}{}, nil }
		timeout = time.Millisecond
		retry   = loadbalancer.Retry(999, timeout, loadbalancer.NewRoundRobin(sd.FixedSubscriber{0: e}))
		errs    = make(chan error, 1)
		invoke  = func() { _, err := retry(context.Background(), struct{}{}); errs <- err }
	)

	go func() { step <- struct{}{} }() // queue up a flush of the endpoint
	invoke()                           // invoke the endpoint and trigger the flush
	if err := <-errs; err != nil {     // that should succeed
		t.Error(err)
	}

	go func() { time.Sleep(10 * timeout); step <- struct{}{} }() // a delayed flush
	invoke()                                                     // invoke the endpoint
	if err := <-errs; err != context.DeadlineExceeded {          // that should not succeed
		t.Errorf("wanted %v, got none", context.DeadlineExceeded)
	}
}
