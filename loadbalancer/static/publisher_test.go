package static_test

import (
	"fmt"
	"io"
	"testing"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer/static"
	"github.com/go-kit/kit/log"
)

func TestStatic(t *testing.T) {
	var (
		instances = []string{"foo", "bar", "baz"}
		endpoints = map[string]endpoint.Endpoint{
			"foo": func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil },
			"bar": func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil },
			"baz": func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil },
		}
		factory = func(instance string) (endpoint.Endpoint, io.Closer, error) {
			if e, ok := endpoints[instance]; ok {
				return e, nil, nil
			}
			return nil, nil, fmt.Errorf("%s: not found", instance)
		}
	)
	p := static.NewPublisher(instances, factory, log.NewNopLogger())
	have, err := p.Endpoints()
	if err != nil {
		t.Fatal(err)
	}
	want := []endpoint.Endpoint{endpoints["foo"], endpoints["bar"], endpoints["baz"]}
	if fmt.Sprint(want) != fmt.Sprint(have) {
		t.Fatalf("want %v, have %v", want, have)
	}
}
