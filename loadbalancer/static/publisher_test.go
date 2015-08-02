package static_test

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/loadbalancer/static"
)

func TestStatic(t *testing.T) {
	var (
		e1        = func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil }
		e2        = func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil }
		endpoints = []endpoint.Endpoint{e1, e2}
	)
	p := static.Publisher(endpoints)
	have, err := p.Endpoints()
	if err != nil {
		t.Fatal(err)
	}
	if want := endpoints; !reflect.DeepEqual(want, have) {
		t.Fatalf("want %#+v, have %#+v", want, have)
	}
}
