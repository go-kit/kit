package loadbalancer_test

import (
	"errors"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer"
	"github.com/go-kit/kit/loadbalancer/fixed"
)

func TestRetryMaxTotalFail(t *testing.T) {
	var (
		endpoints = []endpoint.Endpoint{} // no endpoints
		p         = fixed.NewPublisher(endpoints)
		lb        = loadbalancer.NewRoundRobin(p)
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
		retries = len(endpoints) - 1 // not quite enough retries
		p       = fixed.NewPublisher(endpoints)
		lb      = loadbalancer.NewRoundRobin(p)
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
		retries = len(endpoints) // exactly enough retries
		p       = fixed.NewPublisher(endpoints)
		lb      = loadbalancer.NewRoundRobin(p)
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
		retry   = loadbalancer.Retry(999, timeout, loadbalancer.NewRoundRobin(fixed.NewPublisher([]endpoint.Endpoint{e})))
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
