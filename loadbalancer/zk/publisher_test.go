package zk

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/eapache/channels"
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer"
	"github.com/go-kit/kit/log"
)

var (
	path   = "/gokit.test/service.name"
	e      = func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil }
	logger = log.NewNopLogger()
)

func TestPublisher(t *testing.T) {
	client := newFakeClient()

	p, err := NewPublisher(client, path, NewFactory(""), logger)
	if err != nil {
		t.Fatalf("failed to create new publisher: %v", err)
	}
	defer p.Stop()

	if _, err := p.Endpoints(); err != nil {
		t.Fatal(err)
	}
}

func TestBadFactory(t *testing.T) {
	client := newFakeClient()

	p, err := NewPublisher(client, path, NewFactory("kaboom"), logger)
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

func TestServiceUpdate(t *testing.T) {
	client := newFakeClient()

	p, err := NewPublisher(client, path, NewFactory(""), logger)
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

	// instance1 came online
	client.AddService(path+"/instance1", "zookeeper_node_data")

	// test if we received the instance
	endpoints, err = p.Endpoints()
	if err != nil {
		t.Fatal(err)
	}
	if want, have := 1, len(endpoints); want != have {
		t.Errorf("want %q, have %q", want, have)
	}

	// instance2 came online
	client.AddService(path+"/instance2", "zookeeper_node_data2")

	// test if we received the instance
	endpoints, err = p.Endpoints()
	if err != nil {
		t.Fatal(err)
	}
	if want, have := 2, len(endpoints); want != have {
		t.Errorf("want %q, have %q", want, have)
	}

	// watch triggers an error...
	client.SendErrorOnWatch()

	// test if we ignored the empty instance response due to the error
	endpoints, err = p.Endpoints()
	if err != nil {
		t.Fatal(err)
	}
	if want, have := 2, len(endpoints); want != have {
		t.Errorf("want %q, have %q", want, have)
	}

	// instances go offline
	client.RemoveService(path + "/instance1")
	client.RemoveService(path + "/instance2")

	endpoints, err = p.Endpoints()
	if err != nil {
		t.Fatal(err)
	}
	if want, have := 0, len(endpoints); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}

type fakeClient struct {
	ch        chan bool
	responses map[string]string
	result    bool
}

func newFakeClient() *fakeClient {
	return &fakeClient{
		make(chan bool, 1),
		make(map[string]string),
		true,
	}
}

func (c *fakeClient) CreateParentNodes(path string) error {
	return nil
}

func (c *fakeClient) GetEntries(path string) ([]string, channels.SimpleOutChannel, error) {
	responses := []string{}
	if c.result == false {
		c.result = true
		return responses, channels.Wrap(c.ch), errors.New("Dummy Error")
	}
	for _, data := range c.responses {
		responses = append(responses, data)
	}
	return responses, channels.Wrap(c.ch), nil
}

func (c *fakeClient) AddService(node, data string) {
	c.responses[node] = data
	c.triggerWatch()
}

func (c *fakeClient) RemoveService(node string) {
	delete(c.responses, node)
	c.triggerWatch()
}

func (c *fakeClient) SendErrorOnWatch() {
	c.result = false
	c.triggerWatch()
}

func NewFactory(fakeError string) loadbalancer.Factory {
	return func(string) (endpoint.Endpoint, io.Closer, error) {
		if fakeError == "" {
			return e, nil, nil
		}
		return nil, nil, errors.New(fakeError)
	}
}

func (c *fakeClient) triggerWatch() {
	c.ch <- true
	// watches on ZooKeeper Nodes trigger once, most ZooKeeper libraries also
	// implement "fire once" channels for these watches
	close(c.ch)
	c.ch = make(chan bool, 1)

	// make sure we allow the Publisher to handle this update
	time.Sleep(50 * time.Millisecond)
}
