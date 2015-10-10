package consul_test

import (
	"io"
	"testing"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer/consul"
	"github.com/go-kit/kit/log"
)

func TestPublisher(t *testing.T) {
	var (
		client      = fakeClient{"foo:123", "bar:456"}
		service     = "service"
		tag         = "tag"
		e           = func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil }
		factory     = func(string) (endpoint.Endpoint, io.Closer, error) { return e, nil, nil }
		logger      = log.NewNopLogger()
		passingOnly = true
		publisher   = consul.NewPublisher(client, service, tag, passingOnly, factory, logger)
	)
	defer publisher.Stop()
	endpoints, err := publisher.Endpoints()
	if err != nil {
		t.Fatal(err)
	}
	if want, have := len(client), len(endpoints); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

type fakeClient []string

func (c fakeClient) Service(string, string, bool, uint64) ([]string, uint64, error) {
	return []string(c), 0, nil
}
