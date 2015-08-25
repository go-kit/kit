package fixed_test

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer/fixed"
)

func TestFixed(t *testing.T) {
	var (
		e1        = func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil }
		e2        = func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil }
		endpoints = []endpoint.Endpoint{e1, e2}
	)
	p := fixed.NewPublisher(endpoints)
	have, err := p.Endpoints()
	if err != nil {
		t.Fatal(err)
	}
	if want := endpoints; !reflect.DeepEqual(want, have) {
		t.Fatalf("want %#+v, have %#+v", want, have)
	}
}

func TestFixedReplace(t *testing.T) {
	p := fixed.NewPublisher([]endpoint.Endpoint{
		func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil },
	})
	have, err := p.Endpoints()
	if err != nil {
		t.Fatal(err)
	}
	if want, have := 1, len(have); want != have {
		t.Fatalf("want %d, have %d", want, have)
	}
	p.Replace([]endpoint.Endpoint{})
	have, err = p.Endpoints()
	if err != nil {
		t.Fatal(err)
	}
	if want, have := 0, len(have); want != have {
		t.Fatalf("want %d, have %d", want, have)
	}
}
