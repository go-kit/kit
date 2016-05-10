package etcd_test

import (
	"errors"
	"io"
	"testing"

	stdetcd "github.com/coreos/etcd/client"
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	kitetcd "github.com/go-kit/kit/loadbalancer/etcd"
	"github.com/go-kit/kit/log"
)

var (
	node = &stdetcd.Node{
		Key: "/foo",
		Nodes: []*stdetcd.Node{
			{Key: "/foo/1", Value: "1:1"},
			{Key: "/foo/2", Value: "1:2"},
		},
	}
	fakeResponse = &stdetcd.Response{
		Node: node,
	}
)

func TestPublisher(t *testing.T) {
	var (
		logger = log.NewNopLogger()
		e      = func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil }
	)

	factory := func(string) (endpoint.Endpoint, io.Closer, error) {
		return e, nil, nil
	}

	client := &fakeClient{
		responses: map[string]*stdetcd.Response{"/foo": fakeResponse},
	}

	p, err := kitetcd.NewPublisher(client, "/foo", factory, logger)
	if err != nil {
		t.Fatalf("failed to create new publisher: %v", err)
	}
	defer p.Stop()

	if _, err := p.Endpoints(); err != nil {
		t.Fatal(err)
	}
}

func TestBadFactory(t *testing.T) {
	logger := log.NewNopLogger()

	factory := func(string) (endpoint.Endpoint, io.Closer, error) {
		return nil, nil, errors.New("kaboom")
	}

	client := &fakeClient{
		responses: map[string]*stdetcd.Response{"/foo": fakeResponse},
	}

	p, err := kitetcd.NewPublisher(client, "/foo", factory, logger)
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

type fakeClient struct {
	responses map[string]*stdetcd.Response
}

func (c *fakeClient) GetEntries(prefix string) ([]string, error) {
	response, ok := c.responses[prefix]
	if !ok {
		return nil, errors.New("key not exist")
	}

	entries := make([]string, len(response.Node.Nodes))
	for i, node := range response.Node.Nodes {
		entries[i] = node.Value
	}
	return entries, nil
}

func (c *fakeClient) WatchPrefix(prefix string, responseChan chan *stdetcd.Response) {}
