package consul

import (
	"io"
	"testing"

	consul "github.com/hashicorp/consul/api"
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

var consulState = []*consul.ServiceEntry{
	{
		Node: &consul.Node{
			Address: "10.0.0.0",
			Node:    "app00.local",
		},
		Service: &consul.AgentService{
			ID:      "search-api-0",
			Port:    8000,
			Service: "search",
			Tags: []string{
				"api",
				"v1",
			},
		},
	},
	{
		Node: &consul.Node{
			Address: "10.0.0.1",
			Node:    "app01.local",
		},
		Service: &consul.AgentService{
			ID:      "search-api-1",
			Port:    8001,
			Service: "search",
			Tags: []string{
				"api",
				"v2",
			},
		},
	},
	{
		Node: &consul.Node{
			Address: "10.0.0.1",
			Node:    "app01.local",
		},
		Service: &consul.AgentService{
			Address: "10.0.0.10",
			ID:      "search-db-0",
			Port:    9000,
			Service: "search",
			Tags: []string{
				"db",
			},
		},
	},
}

func TestPublisher(t *testing.T) {
	var (
		logger = log.NewNopLogger()
		client = newTestClient(consulState)
	)

	p, err := NewPublisher(client, testFactory, logger, "search", "api")
	if err != nil {
		t.Fatalf("publisher setup failed: %s", err)
	}
	defer p.Stop()

	eps, err := p.Endpoints()
	if err != nil {
		t.Fatalf("endpoints failed: %s", err)
	}

	if have, want := len(eps), 2; have != want {
		t.Errorf("have %v, want %v", have, want)
	}
}

func TestPublisherNoService(t *testing.T) {
	var (
		logger = log.NewNopLogger()
		client = newTestClient(consulState)
	)

	p, err := NewPublisher(client, testFactory, logger, "feed")
	if err != nil {
		t.Fatalf("publisher setup failed: %s", err)
	}
	defer p.Stop()

	eps, err := p.Endpoints()
	if err != nil {
		t.Fatalf("endpoints failed: %s", err)
	}

	if have, want := len(eps), 0; have != want {
		t.Fatalf("have %v, want %v", have, want)
	}
}

func TestPublisherWithTags(t *testing.T) {
	var (
		logger = log.NewNopLogger()
		client = newTestClient(consulState)
	)

	p, err := NewPublisher(client, testFactory, logger, "search", "api", "v2")
	if err != nil {
		t.Fatalf("publisher setup failed: %s", err)
	}
	defer p.Stop()

	eps, err := p.Endpoints()
	if err != nil {
		t.Fatalf("endpoints failed: %s", err)
	}

	if have, want := len(eps), 1; have != want {
		t.Fatalf("have %v, want %v", have, want)
	}
}

func TestPublisherAddressOverride(t *testing.T) {
	var (
		ctx    = context.Background()
		logger = log.NewNopLogger()
		client = newTestClient(consulState)
	)

	p, err := NewPublisher(client, testFactory, logger, "search", "db")
	if err != nil {
		t.Fatalf("publisher setup failed: %s", err)
	}
	defer p.Stop()

	eps, err := p.Endpoints()
	if err != nil {
		t.Fatalf("endpoints failed: %s", err)
	}

	if have, want := len(eps), 1; have != want {
		t.Fatalf("have %v, want %v", have, want)
	}

	ins, err := eps[0](ctx, struct{}{})
	if err != nil {
		t.Fatal(err)
	}

	if have, want := ins.(string), "10.0.0.10:9000"; have != want {
		t.Errorf("have %#v, want %#v", have, want)
	}
}

type testClient struct {
	entries []*consul.ServiceEntry
}

func newTestClient(entries []*consul.ServiceEntry) Client {
	if entries == nil {
		entries = []*consul.ServiceEntry{}
	}

	return &testClient{
		entries: entries,
	}
}

func (c *testClient) Service(
	service string,
	tag string,
	opts *consul.QueryOptions,
) ([]*consul.ServiceEntry, *consul.QueryMeta, error) {
	es := []*consul.ServiceEntry{}

	for _, e := range c.entries {
		if e.Service.Service != service {
			continue
		}
		if tag != "" {
			tagMap := map[string]struct{}{}

			for _, t := range e.Service.Tags {
				tagMap[t] = struct{}{}
			}

			if _, ok := tagMap[tag]; !ok {
				continue
			}
		}

		es = append(es, e)
	}

	return es, &consul.QueryMeta{}, nil
}

func testFactory(ins string) (endpoint.Endpoint, io.Closer, error) {
	return func(context.Context, interface{}) (interface{}, error) {
		return ins, nil
	}, nil, nil
}
