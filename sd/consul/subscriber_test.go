package consul

import (
	"io"
	"testing"

	consul "github.com/hashicorp/consul/api"
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/service"
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

func TestSubscriber(t *testing.T) {
	var (
		logger = log.NewNopLogger()
		client = newTestClient(consulState)
	)

	s, err := NewSubscriber(client, testFactory, logger, "search", "api")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Stop()

	eps, err := s.Services()
	if err != nil {
		t.Fatal(err)
	}

	if want, have := 2, len(eps); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestSubscriberNoService(t *testing.T) {
	var (
		logger = log.NewNopLogger()
		client = newTestClient(consulState)
	)

	s, err := NewSubscriber(client, testFactory, logger, "feed")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Stop()

	services, err := s.Services()
	if err != nil {
		t.Fatal(err)
	}

	if want, have := 0, len(services); want != have {
		t.Fatalf("want %d, have %d", want, have)
	}
}

func TestSubscriberWithTags(t *testing.T) {
	var (
		logger = log.NewNopLogger()
		client = newTestClient(consulState)
	)

	s, err := NewSubscriber(client, testFactory, logger, "search", "api", "v2")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Stop()

	services, err := s.Services()
	if err != nil {
		t.Fatal(err)
	}

	if want, have := 1, len(services); want != have {
		t.Fatalf("want %d, have %d", want, have)
	}
}

func TestSubscriberAddressOverride(t *testing.T) {
	s, err := NewSubscriber(newTestClient(consulState), testFactory, log.NewNopLogger(), "search", "db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Stop()

	services, err := s.Services()
	if err != nil {
		t.Fatal(err)
	}

	if want, have := 1, len(services); want != have {
		t.Fatalf("want %d, have %d", want, have)
	}

	endpoint, err := services[0].Endpoint("irrelevant")
	if err != nil {
		t.Fatal(err)
	}

	response, err := endpoint(context.Background(), struct{}{})
	if err != nil {
		t.Fatal(err)
	}

	if want, have := "10.0.0.10:9000", response.(string); want != have {
		t.Errorf("want %q, have %q", want, have)
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

func (c *testClient) Service(service, tag string, opts *consul.QueryOptions) ([]*consul.ServiceEntry, *consul.QueryMeta, error) {
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

func testFactory(instance string) (service.Service, io.Closer, error) {
	return service.Func(func(method string) (endpoint.Endpoint, error) {
		return func(context.Context, interface{}) (interface{}, error) {
			return instance, nil
		}, nil
	}), nil, nil
}