package zk

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/samuel/go-zookeeper/zk"
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
