package etcd

import (
	"errors"
	"testing"

	"github.com/coreos/go-etcd/etcd"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer"
	"github.com/go-kit/kit/log"
	"golang.org/x/net/context"
)

var (
	node = &etcd.Node{
		Key: "/foo",
		Nodes: []*etcd.Node{
			{Key: "/foo/1", Value: "1:1"},
			{Key: "/foo/2", Value: "1:2"},
		},
	}
	fakeResponse = &etcd.Response{
		Node: node,
	}
)

func TestPublisher(t *testing.T) {
	var (
		logger = log.NewNopLogger()
		e      = func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil }
	)

	factory := func(instance string) (endpoint.Endpoint, error) {
		return e, nil
	}

	client := &FakeEtcdClient{
		responses: map[string]*etcd.Response{"/foo": fakeResponse},
	}

	p, err := NewPublisher(client, "/foo", factory, logger)
	if err != nil {
		t.Fatalf("failed to create new publisher: %v", err)
	}
	defer p.Stop()

	if _, err := p.Endpoints(); err != nil {
		t.Fatal(err)
	}
}

func TestBadFactory(t *testing.T) {
	var (
		logger = log.NewNopLogger()
		e      = func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil }
	)

	factory := func(instance string) (endpoint.Endpoint, error) {
		return e, errors.New("_")
	}

	client := &FakeEtcdClient{
		responses: map[string]*etcd.Response{"/foo": fakeResponse},
	}

	p, err := NewPublisher(client, "/foo", factory, logger)
	if err != nil {
		t.Fatalf("failed to create new publisher: %v", err)
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

func TestPublisherStoppped(t *testing.T) {
	var (
		logger = log.NewNopLogger()
		e      = func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil }
	)

	factory := func(instance string) (endpoint.Endpoint, error) {
		return e, errors.New("_")
	}

	client := &FakeEtcdClient{
		responses: map[string]*etcd.Response{"/foo": fakeResponse},
	}

	p, err := NewPublisher(client, "/foo", factory, logger)
	if err != nil {
		t.Fatalf("failed to create new publisher: %v", err)
	}

	p.Stop()

	_, have := p.Endpoints()
	if want := loadbalancer.ErrPublisherStopped; want != have {
		t.Fatalf("want %v, have %v", want, have)
	}
}
