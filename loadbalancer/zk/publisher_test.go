package zk

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer"
	"github.com/go-kit/kit/log"
	"github.com/samuel/go-zookeeper/zk"
)

var (
	path   = "/gokit.test/service.name"
	e      = func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil }
	logger = log.NewNopLogger()
)

func TestPublisher(t *testing.T) {
	client := newFakeClient()

	p, err := NewPublisher(client, path, newFactory(""), logger)
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

	p, err := NewPublisher(client, path, newFactory("kaboom"), logger)
	if err != nil {
		t.Fatalf("failed to create new publisher: %v", err)
	}
	defer p.Stop()

	// instance1 came online
	client.AddService(path+"/instance1", "kaboom")

	// instance2 came online
	client.AddService(path+"/instance2", "zookeeper_node_data")

	if err = asyncTest(100*time.Millisecond, 1, p); err != nil {
		t.Error(err)
	}
}

func TestServiceUpdate(t *testing.T) {
	client := newFakeClient()

	p, err := NewPublisher(client, path, newFactory(""), logger)
	if err != nil {
		t.Fatalf("failed to create new publisher: %v", err)
	}
	defer p.Stop()

	endpoints, err := p.Endpoints()
	if err != nil {
		t.Fatal(err)
	}

	if want, have := 0, len(endpoints); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	// instance1 came online
	client.AddService(path+"/instance1", "zookeeper_node_data")

	// instance2 came online
	client.AddService(path+"/instance2", "zookeeper_node_data2")

	// we should have 2 instances
	if err = asyncTest(100*time.Millisecond, 2, p); err != nil {
		t.Error(err)
	}

	// watch triggers an error...
	client.SendErrorOnWatch()

	// test if error was consumed
	if err = client.ErrorIsConsumed(100 * time.Millisecond); err != nil {
		t.Error(err)
	}

	// instance3 came online
	client.AddService(path+"/instance3", "zookeeper_node_data3")

	// we should have 3 instances
	if err = asyncTest(100*time.Millisecond, 3, p); err != nil {
		t.Error(err)
	}

	// instance1 goes offline
	client.RemoveService(path + "/instance1")

	// instance2 goes offline
	client.RemoveService(path + "/instance2")

	// we should have 1 instance
	if err = asyncTest(100*time.Millisecond, 1, p); err != nil {
		t.Error(err)
	}
}

func TestBadPublisherCreate(t *testing.T) {
	client := newFakeClient()
	client.SendErrorOnWatch()
	p, err := NewPublisher(client, path, newFactory(""), logger)
	if err == nil {
		t.Error("expected error on new publisher")
	}
	if p != nil {
		t.Error("expected publisher not to be created")
	}
	p, err = NewPublisher(client, "BadPath", newFactory(""), logger)
	if err == nil {
		t.Error("expected error on new publisher")
	}
	if p != nil {
		t.Error("expected publisher not to be created")
	}
}

type fakeClient struct {
	mtx       sync.Mutex
	ch        chan zk.Event
	responses map[string]string
	result    bool
}

func newFakeClient() *fakeClient {
	return &fakeClient{
		ch:        make(chan zk.Event, 5),
		responses: make(map[string]string),
		result:    true,
	}
}

func (c *fakeClient) CreateParentNodes(path string) error {
	if path == "BadPath" {
		return errors.New("Dummy Error")
	}
	return nil
}

func (c *fakeClient) GetEntries(path string) ([]string, <-chan zk.Event, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if c.result == false {
		c.result = true
		return []string{}, c.ch, errors.New("Dummy Error")
	}
	responses := []string{}
	for _, data := range c.responses {
		responses = append(responses, data)
	}
	return responses, c.ch, nil
}

func (c *fakeClient) AddService(node, data string) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.responses[node] = data
	c.ch <- zk.Event{}
}

func (c *fakeClient) RemoveService(node string) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	delete(c.responses, node)
	c.ch <- zk.Event{}
}

func (c *fakeClient) SendErrorOnWatch() {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.result = false
	c.ch <- zk.Event{}
}

func (c *fakeClient) ErrorIsConsumed(t time.Duration) error {
	timeout := time.After(t)
	for {
		select {
		case <-timeout:
			return fmt.Errorf("expected error not consumed after timeout %s", t.String())
		default:
			c.mtx.Lock()
			if c.result == false {
				c.mtx.Unlock()
				return nil
			}
			c.mtx.Unlock()
		}
	}
}

func (c *fakeClient) Stop() {}

func newFactory(fakeError string) loadbalancer.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		if fakeError == instance {
			return nil, nil, errors.New(fakeError)
		}
		return e, nil, nil
	}
}

func asyncTest(timeout time.Duration, want int, p *Publisher) (err error) {
	var endpoints []endpoint.Endpoint
	// want can never be -1
	have := -1
	t := time.After(timeout)
	for {
		select {
		case <-t:
			return fmt.Errorf("want %d, have %d after timeout %s", want, have, timeout.String())
		default:
			endpoints, err = p.Endpoints()
			have = len(endpoints)
			if err != nil || want == have {
				return
			}
			time.Sleep(time.Millisecond)
		}
	}
}
