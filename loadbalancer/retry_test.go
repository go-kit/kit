package loadbalancer_test

import (
	"errors"
	"runtime"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer"
	"golang.org/x/net/context"

	"testing"
)

func TestRetry(t *testing.T) {
	var (
		endpoints = []endpoint.Endpoint{}
		p         = loadbalancer.NewStaticPublisher(endpoints)
		lb        = loadbalancer.RoundRobin(p)
	)

	{
		max := 999
		e := loadbalancer.Retry(max, lb)
		if _, err := e(context.Background(), struct{}{}); err == nil {
			t.Errorf("expected error, got none")
		}
	}

	endpoints = []endpoint.Endpoint{
		func(context.Context, interface{}) (interface{}, error) { return nil, errors.New("error one") },
		func(context.Context, interface{}) (interface{}, error) { return nil, errors.New("error two") },
		func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil /* OK */ },
	}
	p.Replace(endpoints)
	runtime.Gosched()

	{
		max := len(endpoints) - 1
		e := loadbalancer.Retry(max, lb)
		if _, err := e(context.Background(), struct{}{}); err == nil {
			t.Errorf("expected error, got none")
		}
	}

	{
		max := len(endpoints)
		e := loadbalancer.Retry(max, lb)
		if _, err := e(context.Background(), struct{}{}); err != nil {
			t.Error(err)
		}
	}
}
