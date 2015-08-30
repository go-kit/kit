package appdash

import (
	"github.com/go-kit/kit/endpoint"
	"golang.org/x/net/context"
	"reflect"
	"sourcegraph.com/sourcegraph/appdash"
	"testing"
	"time"
)

func TestMiddlewareWithDefaultEndpoint(t *testing.T) {
	ms := appdash.NewMemoryStore()
	c := appdash.NewLocalCollector(ms)

	newEventFunc := NewDefaultEndpointEventFunc("TEST")

	ctx := context.Background()

	spanID := appdash.SpanID{1, 2, 3} // Trace,Span,Parent
	ctx = context.WithValue(ctx, SpanContextKey, &spanID)

	var e endpoint.Endpoint
	e = func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil }
	e = NewTrace(newEventFunc, c)(e)
	e(ctx, struct{}{})

	trace, err := ms.Trace(1)
	if err != nil {
		t.Fatal(err)
	}

	var event DefaultEndpointEvent
	if err := appdash.UnmarshalEvent(trace.Span.Annotations, &event); err != nil {
		t.Fatal(err)
	}

	wantEvent := DefaultEndpointEvent{
		Name: "TEST",
	}
	event.Recv = time.Time{}
	event.Send = time.Time{}

	if !reflect.DeepEqual(event, wantEvent) {
		t.Errorf("got EndpointEvent %+v,want %+v", event, wantEvent)
	}
}
