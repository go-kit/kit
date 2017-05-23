package consul

import (
	"context"
	"errors"
	"io"
	"reflect"
	"testing"

	stdconsul "github.com/hashicorp/consul/api"

	"github.com/go-kit/kit/endpoint"
)

func TestClientRegistration(t *testing.T) {
	c := newTestClient(nil, nil)

	services, _, err := c.Service(testRegistration.Name, "", true, &stdconsul.QueryOptions{})
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

	services, _, err = c.Service(testRegistration.Name, "", true, &stdconsul.QueryOptions{})
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

	services, _, err = c.Service(testRegistration.Name, "", true, &stdconsul.QueryOptions{})
	if err != nil {
		t.Error(err)
	}
	if want, have := 0, len(services); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	if err := c.CheckRegister(testCheck); err != nil {
		t.Error(err)
	}
	if err := c.UpdateTTL(testCheckID, "", "pass"); err != nil {
		t.Error(err)
	}

	checks, _, err := c.Checks(testRegistration.ID, &stdconsul.QueryOptions{})
	if err != nil {
		t.Error(err)
	}
	if want, have := 1, len(checks); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	if err := c.CheckDeregister(testCheckID); err != nil {
		t.Error(err)
	}

	checks, _, err = c.Checks(testRegistration.ID, &stdconsul.QueryOptions{})
	if err != nil {
		t.Error(err)
	}
	if want, have := 0, len(checks); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

type testClient struct {
	entries []*stdconsul.ServiceEntry
	checks  []*stdconsul.AgentCheckRegistration
}

func newTestClient(entries []*stdconsul.ServiceEntry, checks []*stdconsul.AgentCheckRegistration) *testClient {
	return &testClient{
		entries: entries,
		checks:  checks,
	}
}

var _ Client = &testClient{}

func (c *testClient) Service(service, tag string, _ bool, opts *stdconsul.QueryOptions) ([]*stdconsul.ServiceEntry, *stdconsul.QueryMeta, error) {
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

func (c *testClient) Register(r *stdconsul.AgentServiceRegistration) error {
	toAdd := registration2entry(r)

	for _, entry := range c.entries {
		if reflect.DeepEqual(*entry, *toAdd) {
			return errors.New("duplicate")
		}
	}

	c.entries = append(c.entries, toAdd)
	return nil
}

func (c *testClient) Deregister(r *stdconsul.AgentServiceRegistration) error {
	toDelete := registration2entry(r)

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

func (c *testClient) CheckRegister(ac *stdconsul.AgentCheckRegistration) error {
	for _, check := range c.checks {
		if reflect.DeepEqual(*check, *ac) {
			return errors.New("duplicate")
		}
	}

	c.checks = append(c.checks, ac)
	return nil

}

func (c *testClient) CheckDeregister(id string) error {
	var newChecks []*stdconsul.AgentCheckRegistration
	for _, check := range c.checks {
		if check.ID == id {
			continue
		}
		newChecks = append(newChecks, check)
	}
	if len(newChecks) == len(c.checks) {
		return errors.New("not found")
	}

	c.checks = newChecks
	return nil
}

func (c *testClient) Checks(service string, opts *stdconsul.QueryOptions) (stdconsul.HealthChecks, *stdconsul.QueryMeta, error) {
	var results stdconsul.HealthChecks

	for _, check := range c.checks {
		if check.ServiceID != service {
			continue
		}
		results = append(results, &stdconsul.HealthCheck{
			ServiceID: service,
			CheckID:   check.ID,
		})
	}

	return results, &stdconsul.QueryMeta{}, nil
}

func (c *testClient) UpdateTTL(checkID, output, status string) error {
	for _, check := range c.checks {
		if check.ID == checkID {
			return nil
		}
	}
	return errors.New("not found")

}

func registration2entry(r *stdconsul.AgentServiceRegistration) *stdconsul.ServiceEntry {
	return &stdconsul.ServiceEntry{
		Node: &stdconsul.Node{
			Node:    "some-node",
			Address: r.Address,
		},
		Service: &stdconsul.AgentService{
			ID:      r.ID,
			Service: r.Name,
			Tags:    r.Tags,
			Port:    r.Port,
			Address: r.Address,
		},
		// Checks ignored
	}
}

func testFactory(instance string) (endpoint.Endpoint, io.Closer, error) {
	return func(context.Context, interface{}) (interface{}, error) {
		return instance, nil
	}, nil, nil
}

var testRegistration = &stdconsul.AgentServiceRegistration{
	ID:      "my-id",
	Name:    "my-name",
	Tags:    []string{"my-tag-1", "my-tag-2"},
	Port:    12345,
	Address: "my-address",
}

var testCheckID = "my-id"

var testCheck = &stdconsul.AgentCheckRegistration{
	ID:                testCheckID,
	Name:              "my-name",
	ServiceID:         "my-id",
	AgentServiceCheck: stdconsul.AgentServiceCheck{TTL: "5s"},
}
