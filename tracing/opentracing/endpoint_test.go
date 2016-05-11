package opentracing_test

import (
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	kitot "github.com/go-kit/kit/tracing/opentracing"
)

func TestTraceServer(t *testing.T) {
	tracer := mocktracer.New()

	// Initialize the ctx with a nameless Span.
	contextSpan := tracer.StartSpan("").(*mocktracer.MockSpan)
	ctx := opentracing.ContextWithSpan(context.Background(), contextSpan)

	var innerEndpoint endpoint.Endpoint
	innerEndpoint = func(context.Context, interface{}) (interface{}, error) {
		return struct{}{}, nil
	}
	tracedEndpoint := kitot.TraceServer(tracer, "testOp")(innerEndpoint)
	if _, err := tracedEndpoint(ctx, struct{}{}); err != nil {
		t.Fatal(err)
	}
	if want, have := 1, len(tracer.FinishedSpans); want != have {
		t.Fatalf("Want %v span(s), found %v", want, have)
	}

	endpointSpan := tracer.FinishedSpans[0]
	// Test that the op name is updated
	if want, have := "testOp", endpointSpan.OperationName; want != have {
		t.Fatalf("Want %q, have %q", want, have)
	}
	// ... and that the ID is unmodified.
	if want, have := contextSpan.SpanID, endpointSpan.SpanID; want != have {
		t.Errorf("Want SpanID %q, have %q", want, have)
	}
}

func TestTraceServerNoContextSpan(t *testing.T) {
	tracer := mocktracer.New()

	var innerEndpoint endpoint.Endpoint
	innerEndpoint = func(context.Context, interface{}) (interface{}, error) {
		return struct{}{}, nil
	}
	tracedEndpoint := kitot.TraceServer(tracer, "testOp")(innerEndpoint)
	// Empty/background context:
	if _, err := tracedEndpoint(context.Background(), struct{}{}); err != nil {
		t.Fatal(err)
	}
	// tracedEndpoint created a new Span:
	if want, have := 1, len(tracer.FinishedSpans); want != have {
		t.Fatalf("Want %v span(s), found %v", want, have)
	}

	endpointSpan := tracer.FinishedSpans[0]
	if want, have := "testOp", endpointSpan.OperationName; want != have {
		t.Fatalf("Want %q, have %q", want, have)
	}
}

func TestTraceClient(t *testing.T) {
	tracer := mocktracer.New()

	// Initialize the ctx with a parent Span.
	parentSpan := tracer.StartSpan("parent").(*mocktracer.MockSpan)
	defer parentSpan.Finish()
	ctx := opentracing.ContextWithSpan(context.Background(), parentSpan)

	var innerEndpoint endpoint.Endpoint
	innerEndpoint = func(context.Context, interface{}) (interface{}, error) {
		return struct{}{}, nil
	}
	tracedEndpoint := kitot.TraceClient(tracer, "testOp")(innerEndpoint)
	if _, err := tracedEndpoint(ctx, struct{}{}); err != nil {
		t.Fatal(err)
	}
	// tracedEndpoint created a new Span:
	if want, have := 1, len(tracer.FinishedSpans); want != have {
		t.Fatalf("Want %v span(s), found %v", want, have)
	}

	endpointSpan := tracer.FinishedSpans[0]
	if want, have := "testOp", endpointSpan.OperationName; want != have {
		t.Fatalf("Want %q, have %q", want, have)
	}
	// ... and that the parent ID is set appropriately.
	if want, have := parentSpan.SpanID, endpointSpan.ParentID; want != have {
		t.Errorf("Want ParentID %q, have %q", want, have)
	}
}
