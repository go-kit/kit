// +build integration

package consul

import (
	"io"
	"os"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/service"
	stdconsul "github.com/hashicorp/consul/api"
)

func TestIntegration(t *testing.T) {
	consulAddr := os.Getenv("CONSUL_ADDRESS")
	if consulAddr == "" {
		t.Fatal("CONSUL_ADDRESS is not set")
	}

	stdClient, err := stdconsul.NewClient(&stdconsul.Config{
		Address: consulAddr,
	})
	if err != nil {
		t.Fatal(err)
	}

	var (
		node        = "my-node"
		address     = "my-address:12345"
		datacenter  = "dc1" // by default!
		serviceID   = "my-service-ID"
		serviceName = "my-service"
		tags        = []string{} //"my-tag-1", "my-tag-2"}
		port        = 12345
		checkID     = "my-check-ID"
		name        = "my-check-name"
		status      = "my-status"
		notes       = "my-notes"
		output      = "my-output"
		factory     = func(instance string) (service.Service, io.Closer, error) { return nil, nil, nil }
	)

	client := NewClient(stdClient)
	logger := log.NewLogfmtLogger(os.Stderr)

	publisher := NewPublisher(client, &stdconsul.CatalogRegistration{
		Node:       node,
		Address:    address,
		Datacenter: datacenter,
		Service: &stdconsul.AgentService{
			ID:      serviceID,
			Service: serviceName,
			Tags:    tags,
			Port:    port,
			Address: address,
		},
		Check: &stdconsul.AgentCheck{
			Node:        node,
			CheckID:     checkID,
			Name:        name,
			Status:      status,
			Notes:       notes,
			Output:      output,
			ServiceID:   serviceID,
			ServiceName: serviceName,
		},
	}, log.NewContext(logger).With("component", "publisher"))

	publisher.Publish()
	//defer publisher.Unpublish()
	time.Sleep(time.Second)

	subscriber, err := NewSubscriber(client, factory, log.NewContext(logger).With("component", "subscriber"), serviceName, tags...)
	if err != nil {
		t.Fatal(err)
	}

	services, err := subscriber.Services()
	if err != nil {
		t.Error(err)
	}

	if want, have := 1, len(services); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
	if len(services) > 0 {
		t.Logf("%#+v", services[0])
	}
}
