package endpoint_test

import (
	"context"
	"testing"

	"github.com/go-kit/kit/endpoint"
)

func TestEndpointNameMiddleware(t *testing.T) {
	ctx := context.Background()

	var name string

	ep := func(ctx context.Context, request interface{}) (interface{}, error) {
		name = ctx.Value(endpoint.ContextKeyEndpointName).(string)

		return nil, nil
	}

	mw := endpoint.EndpointNameMiddleware("go-kit/endpoint")

	mw(ep)(ctx, nil)

	if want, have := "go-kit/endpoint", name; want != have {
		t.Fatalf("unexpected endpoint name, wanted %q, got %q", want, have)
	}
}
