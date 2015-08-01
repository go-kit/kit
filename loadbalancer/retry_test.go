package loadbalancer_test

import (
	"errors"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer"
	"github.com/go-kit/kit/loadbalancer/publisher/static"
	"github.com/go-kit/kit/loadbalancer/strategy"
)

func TestRetryMax(t *testing.T) {
	var (
		endpoints = []endpoint.Endpoint{}
		p         = static.NewPublisher(endpoints)
		lb        = strategy.RoundRobin(p)
	)

	if _, err := loadbalancer.Retry(999, time.Second, lb)(context.Background(), struct{}{}); err == nil {
		t.Errorf("expected error, got none")
	}

	endpoints = []endpoint.Endpoint{
		func(context.Context, interface{}) (interface{}, error) { return nil, errors.New("error one") },
		func(context.Context, interface{}) (interface{}, error) { return nil, errors.New("error two") },
		func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil /* OK */ },
	}
	p.Replace(endpoints)
	time.Sleep(10 * time.Millisecond) //assertLoadBalancerNotEmpty(t, lb) // TODO

	if _, err := loadbalancer.Retry(len(endpoints)-1, time.Second, lb)(context.Background(), struct{}{}); err == nil {
		t.Errorf("expected error, got none")
	}

	if _, err := loadbalancer.Retry(len(endpoints), time.Second, lb)(context.Background(), struct{}{}); err != nil {
		t.Error(err)
	}
}

func TestRetryTimeout(t *testing.T) {
	var (
		step    = make(chan struct{})
		e       = func(context.Context, interface{}) (interface{}, error) { <-step; return struct{}{}, nil }
		timeout = time.Millisecond
		retry   = loadbalancer.Retry(999, timeout, strategy.RoundRobin(static.NewPublisher([]endpoint.Endpoint{e})))
		errs    = make(chan error)
		invoke  = func() { _, err := retry(context.Background(), struct{}{}); errs <- err }
	)

	go invoke()                    // invoke the endpoint
	step <- struct{}{}             // tell the endpoint to return
	if err := <-errs; err != nil { // that should succeed
		t.Error(err)
	}

	go invoke()                                         // invoke the endpoint
	time.Sleep(2 * timeout)                             // wait
	time.Sleep(2 * timeout)                             // wait again (CI servers!!)
	step <- struct{}{}                                  // tell the endpoint to return
	if err := <-errs; err != context.DeadlineExceeded { // that should not succeed
		t.Errorf("wanted error, got none")
	}
}
