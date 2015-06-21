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

	if _, err := loadbalancer.Retry(999, lb)(context.Background(), struct{}{}); err == nil {
		t.Errorf("expected error, got none")
	}

	endpoints = []endpoint.Endpoint{
		func(context.Context, interface{}) (interface{}, error) { return nil, errors.New("error one") },
		func(context.Context, interface{}) (interface{}, error) { return nil, errors.New("error two") },
		func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil /* OK */ },
	}
	p.Replace(endpoints)
	runtime.Gosched()

	if _, err := loadbalancer.Retry(len(endpoints)-1, lb)(context.Background(), struct{}{}); err == nil {
		t.Errorf("expected error, got none")
	}

	if _, err := loadbalancer.Retry(len(endpoints), lb)(context.Background(), struct{}{}); err != nil {
		t.Error(err)
	}
}
