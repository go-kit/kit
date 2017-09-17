package consul

import (
	"context"
	"testing"

	consul "github.com/hashicorp/consul/api"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
)

var _ sd.Instancer = (*Instancer)(nil) // API check

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

func TestInstancer(t *testing.T) {
	var (
		logger = log.NewNopLogger()
		client = newTestClient(consulState)
	)

	s := NewInstancer(client, logger, "search", []string{"api"}, true)
	defer s.Stop()

	state := s.cache.State()
	if want, have := 2, len(state.Instances); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestInstancerNoService(t *testing.T) {
	var (
		logger = log.NewNopLogger()
		client = newTestClient(consulState)
	)

	s := NewInstancer(client, logger, "feed", []string{}, true)
	defer s.Stop()

	state := s.cache.State()
	if want, have := 0, len(state.Instances); want != have {
		t.Fatalf("want %d, have %d", want, have)
	}
}

func TestInstancerWithTags(t *testing.T) {
	var (
		logger = log.NewNopLogger()
		client = newTestClient(consulState)
	)

	s := NewInstancer(client, logger, "search", []string{"api", "v2"}, true)
	defer s.Stop()

	state := s.cache.State()
	if want, have := 1, len(state.Instances); want != have {
		t.Fatalf("want %d, have %d", want, have)
	}
}

func TestInstancerAddressOverride(t *testing.T) {
	s := NewInstancer(newTestClient(consulState), log.NewNopLogger(), "search", []string{"db"}, true)
	defer s.Stop()

	state := s.cache.State()
	if want, have := 1, len(state.Instances); want != have {
		t.Fatalf("want %d, have %d", want, have)
	}

	endpoint, closer, err := testFactory(state.Instances[0])
	if err != nil {
		t.Fatal(err)
	}
	if closer != nil {
		defer closer.Close()
	}

	response, err := endpoint(context.Background(), struct{}{})
	if err != nil {
		t.Fatal(err)
	}

	if want, have := "10.0.0.10:9000", response.(string); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}
