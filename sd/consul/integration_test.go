// +build integration

package consul

import (
	"io"
	"os"
	"testing"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	stdconsul "github.com/hashicorp/consul/api"
)

func TestIntegration(t *testing.T) {
	consulAddr := os.Getenv("CONSUL_ADDR")
	if consulAddr == "" {
		t.Fatal("CONSUL_ADDR is not set")
	}
	stdClient, err := stdconsul.NewClient(&stdconsul.Config{
		Address: consulAddr,
	})
	if err != nil {
		t.Fatal(err)
	}
	client := NewClient(stdClient)
	logger := log.NewLogfmtLogger(os.Stderr)

	// Produce a fake service registration.
	r := &stdconsul.AgentServiceRegistration{
		ID:                "my-service-ID",
		Name:              "my-service-name",
		Tags:              []string{"alpha", "beta"},
		Port:              12345,
		Address:           "my-address",
		EnableTagOverride: false,
		// skipping check(s)
	}

	// Build a subscriber on r.Name + r.Tags.
	factory := func(instance string) (endpoint.Endpoint, io.Closer, error) {
		t.Logf("factory invoked for %q", instance)
		return endpoint.Nop, nil, nil
	}
	subscriber := NewSubscriber(
		client,
		factory,
		log.With(logger, "component", "subscriber"),
		r.Name,
		r.Tags,
		true,
	)

	time.Sleep(time.Second)

	// Before we publish, we should have no endpoints.
	endpoints, err := subscriber.Endpoints()
	if err != nil {
		t.Error(err)
	}
	if want, have := 0, len(endpoints); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	// Build a registrar for r.
	registrar := NewRegistrar(client, r, log.With(logger, "component", "registrar"))
	registrar.Register()
	defer registrar.Deregister()

	time.Sleep(time.Second)

	// Now we should have one active endpoints.
	endpoints, err = subscriber.Endpoints()
	if err != nil {
		t.Error(err)
	}
	if want, have := 1, len(endpoints); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	// Add a TTL health check
	c := &stdconsul.AgentCheckRegistration{
		ID:                "my-check-ID",
		ServiceID:         "my-service-ID",
		Name:              "my-check-name",
		AgentServiceCheck: stdconsul.AgentServiceCheck{TTL: "5s"},
	}
	registrar.AddCheck(c, testTTLCheck)

	// We should have a registered check
	checks, _, err := client.Checks("my-service-name", &stdconsul.QueryOptions{})
	if err != nil {
		t.Error(err)
	}
	if want, have := 1, len(checks); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	// Normally we would defer this one right after AddCheck
	registrar.RemoveCheck("my-check-ID")

	// We should have no registered checks
	checks, _, err = client.Checks("my-service-name", &stdconsul.QueryOptions{})
	if err != nil {
		t.Error(err)
	}
	if want, have := 0, len(checks); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

}
