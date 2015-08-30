package appdash

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/appdash"

	"github.com/go-kit/kit/endpoint"
)

var (
	// SpanContextKey TODO(pb)
	SpanContextKey = "Appdash-SpanID"
)

// EndpointEvent TODO(pb)
type EndpointEvent interface {
	appdash.Event
	BeforeRequest(interface{})
	AfterResponse(interface{}, error)
}

// NewEndpointEventFunc TODO(pb)
type NewEndpointEventFunc func() EndpointEvent

// NewTrace returns an endpoint.Middleware that extracts a span ID from the
// context, executes event.BeforeRequest and event.AfterResponse for the
// event, and submits the event to the collector. If no span ID is found in
// the context, a new span ID is generated and inserted.
func NewTrace(newEventFunc NewEndpointEventFunc, collector appdash.Collector) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			spanID, ok := fromContext(ctx)
			if !ok {
				spanID = appdash.NewRootSpanID()
				ctx = context.WithValue(ctx, SpanContextKey, spanID)
			}
			rec := appdash.NewRecorder(spanID, collector)
			event := newEventFunc()
			event.BeforeRequest(request)
			response, err := next(ctx, request)
			event.AfterResponse(response, err)
			rec.Event(event)
			return response, err
		}
	}
}

func fromContext(ctx context.Context) (appdash.SpanID, bool) {
	val := ctx.Value(SpanContextKey)
	if val == nil {
		return appdash.SpanID{}, false
	}
	spanID, ok := val.(appdash.SpanID)
	if !ok {
		panic(SpanContextKey + " value isn't a span object")
	}
	return spanID, true
}
