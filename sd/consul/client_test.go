package consul

import (
	"errors"
	"io"
	"reflect"
	"testing"

	stdconsul "github.com/hashicorp/consul/api"
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/service"
)

type testClient struct {
	entries []*stdconsul.ServiceEntry
}

func newTestClient(entries []*stdconsul.ServiceEntry) Client {
	return &testClient{
		entries: entries,
	}
}

func (c *testClient) Service(service, tag string, opts *stdconsul.QueryOptions) ([]*stdconsul.ServiceEntry, *stdconsul.QueryMeta, error) {
	var results []*stdconsul.ServiceEntry

	for _, entry := range c.entries {
		if entry.Service.Service != service {
			continue
		}
		if tag != "" {
			tagMap := map[string]struct{}{}

			for _, t := range entry.Service.Tags {
				tagMap[t] = struct{}{}
			}

			if _, ok := tagMap[tag]; !ok {
				continue
			}
		}

		results = append(results, entry)
	}

	return results, &stdconsul.QueryMeta{}, nil
}

func (c *testClient) Register(registration *stdconsul.CatalogRegistration) error {
	toAdd := registration2entry(registration)

	for _, entry := range c.entries {
		if reflect.DeepEqual(*entry, *toAdd) {
			return errors.New("duplicate")
		}
	}

	c.entries = append(c.entries, toAdd)
	return nil
}

func (c *testClient) Deregister(registration *stdconsul.CatalogRegistration) error {
	toDelete := registration2entry(registration)

	var newEntries []*stdconsul.ServiceEntry
	for _, entry := range c.entries {
		if reflect.DeepEqual(*entry, *toDelete) {
			continue
		}
		newEntries = append(newEntries, entry)
	}
	if len(newEntries) == len(c.entries) {
		return errors.New("not found")
	}

	c.entries = newEntries
	return nil
}

func registration2entry(registration *stdconsul.CatalogRegistration) *stdconsul.ServiceEntry {
	return &stdconsul.ServiceEntry{
		Node: &stdconsul.Node{
			Node:    registration.Node,
			Address: registration.Address,
		},
		Service: registration.Service,
		// Checks ignored
	}
}

func testFactory(instance string) (service.Service, io.Closer, error) {
	return service.Func(func(method string) (endpoint.Endpoint, error) {
		return func(context.Context, interface{}) (interface{}, error) {
			return instance, nil
		}, nil
	}), nil, nil
}

var testRegistration = &stdconsul.CatalogRegistration{
	Node:       "node",
	Address:    "addr",
	Datacenter: "dc",
	Service: &stdconsul.AgentService{
		ID:      "id",
		Service: "service",
		Tags:    []string{"a", "b"},
		Port:    12345,
		Address: "addr",
	},
}

func TestClientRegistration(t *testing.T) {
	c := newTestClient(nil)

	services, _, err := c.Service(testRegistration.Service.Service, "", &stdconsul.QueryOptions{})
	if err != nil {
		t.Error(err)
	}
	if want, have := 0, len(services); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	if err := c.Register(testRegistration); err != nil {
		t.Error(err)
	}

	if err := c.Register(testRegistration); err == nil {
		t.Errorf("want error, have %v", err)
	}

	services, _, err = c.Service(testRegistration.Service.Service, "", &stdconsul.QueryOptions{})
	if err != nil {
		t.Error(err)
	}
	if want, have := 1, len(services); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	if err := c.Deregister(testRegistration); err != nil {
		t.Error(err)
	}

	if err := c.Deregister(testRegistration); err == nil {
		t.Errorf("want error, have %v", err)
	}

	services, _, err = c.Service(testRegistration.Service.Service, "", &stdconsul.QueryOptions{})
	if err != nil {
		t.Error(err)
	}
	if want, have := 0, len(services); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}
