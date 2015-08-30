package appdash

import (
	"github.com/go-kit/kit/endpoint"
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/appdash"
)

var (
	SpanContextKey = "Appdash-SpanID"
)

type EndpointEvent interface {
	appdash.Event
	BeforeRequest(interface{})
	AfterResponse(interface{}, error)
}

type NewEndpointEventFunc func() EndpointEvent

// NewTrace returns a endpoint.Middleware that extracts a span id from the context,
// executes event.BeforeRequest and event.AfterResponse for event
// and submits the event to the collector. If no span id is found in the context,
// a new span id is generated and inserted.
func NewTrace(newEventFunc NewEndpointEventFunc, collector appdash.Collector) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			spanID, ok := fromContext(ctx)
			if !ok {
				_spanID := appdash.NewRootSpanID()
				spanID = &_spanID
				ctx = context.WithValue(ctx, SpanContextKey, spanID)
			}
			rec := appdash.NewRecorder(*spanID, collector)
			event := newEventFunc()
			event.BeforeRequest(request)
			response, err := next(ctx, request)
			event.AfterResponse(response, err)
			rec.Event(event)
			return response, err
		}
	}
}

func fromContext(ctx context.Context) (*appdash.SpanID, bool) {
	val := ctx.Value(SpanContextKey)
	if val == nil {
		return nil, false
	}

	spanID, ok := val.(*appdash.SpanID)
	if !ok {
		panic(SpanContextKey + " value isn't a span object")
	}
	return spanID, true
}
