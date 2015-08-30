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

// EndpointEventFunc TODO(pb)
type EndpointEventFunc func() EndpointEvent

// AnnotateServer returns an endpoint.Middleware that extracts a span ID from
// the context, executes event.BeforeRequest and event.AfterResponse for the
// event, and submits the event to the collector. If no span ID is found in
// the context, a new span ID is generated and inserted.
func AnnotateServer(newEvent EndpointEventFunc, c appdash.Collector) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			spanID, ok := fromContext(ctx)
			if !ok {
				spanID = appdash.NewRootSpanID()
				ctx = context.WithValue(ctx, SpanContextKey, spanID)
			}
			var (
				rec   = appdash.NewRecorder(spanID, c)
				event = newEvent()
			)
			event.BeforeRequest(request)
			defer func() { event.AfterResponse(response, err); rec.Event(event) }()
			response, err = next(ctx, request)
			return
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
