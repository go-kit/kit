package appdash

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/appdash"

	"github.com/go-kit/kit/endpoint"
)

func TestMiddlewareWithDefaultEndpoint(t *testing.T) {
	var (
		ms           = appdash.NewMemoryStore()
		c            = appdash.NewLocalCollector(ms)
		newEventFunc = NewDefaultEndpointEventFunc("TEST")
		spanID       = appdash.SpanID{Trace: 1, Span: 2, Parent: 3}
		ctx          = context.WithValue(context.Background(), SpanContextKey, spanID)
	)

	// Invoke the endpoint.
	var e endpoint.Endpoint
	e = func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil }
	e = NewTrace(newEventFunc, c)(e)
	if _, err := e(ctx, struct{}{}); err != nil {
		t.Fatal(err)
	}

	// Get the generated trace.
	trace, err := ms.Trace(1)
	if err != nil {
		t.Fatal(err)
	}

	// Get the event from the trace.
	var event DefaultEndpointEvent
	if err := appdash.UnmarshalEvent(trace.Span.Annotations, &event); err != nil {
		t.Fatal(err)
	}

	// It should match what we sent, modulo times.
	if want, have := (DefaultEndpointEvent{
		Name: "TEST",
		Recv: event.Recv,
		Send: event.Send,
		Err:  "",
	}), event; !reflect.DeepEqual(want, have) {
		t.Errorf("want %#v, have %#v", want, have)
	}
}
