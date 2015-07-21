package loadbalancer_test

import (
	"errors"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer"
	"golang.org/x/net/context"

	"testing"
)

func TestRetryMax(t *testing.T) {
	var (
		endpoints = []endpoint.Endpoint{}
		p         = loadbalancer.NewStaticPublisher(endpoints)
		lb        = loadbalancer.RoundRobin(p)
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
	assertLoadBalancerNotEmpty(t, lb)

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
		retry   = loadbalancer.Retry(999, timeout, loadbalancer.RoundRobin(loadbalancer.NewStaticPublisher([]endpoint.Endpoint{e})))
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
