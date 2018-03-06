package zipkin_test

import (
	"context"
	"testing"

	"github.com/go-kit/kit/endpoint"
	zipkin "github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/reporter/recorder"

	kitzipkin "github.com/go-kit/kit/tracing/zipkin"
)

func TestTraceServer(t *testing.T) {
	reporter := recorder.NewReporter()
	defer reporter.Close()
	tracer, _ := zipkin.NewTracer(reporter)

	// Initialize the ctx with a nameless Span.
	contextSpan := tracer.StartSpan("")
	ctx := zipkin.NewContext(context.Background(), contextSpan)

	tracedEndpoint := kitzipkin.TraceServer(tracer, "testOp")(endpoint.Nop)
	if _, err := tracedEndpoint(ctx, struct{}{}); err != nil {
		t.Fatal(err)
	}

	finishedSpans := reporter.Flush()
	if want, have := 1, len(finishedSpans); want != have {
		t.Fatalf("Want %v span(s), found %v", want, have)
	}

	// Test that the op name is updated
	endpointSpan := finishedSpans[0]
	if want, have := "testOp", endpointSpan.Name; want != have {
		t.Fatalf("Want %q, have %q", want, have)
	}
	contextContext := contextSpan.Context()
	endpointContext := endpointSpan.SpanContext
	// ...and that the ID is unmodified.
	if want, have := contextContext.ID, endpointContext.ID; want != have {
		t.Errorf("Want SpanID %q, have %q", want, have)
	}
}

func TestTraceServerNoContextSpan(t *testing.T) {
	reporter := recorder.NewReporter()
	defer reporter.Close()
	tracer, _ := zipkin.NewTracer(reporter)

	// Empty/background context.
	tracedEndpoint := kitzipkin.TraceServer(tracer, "testOp")(endpoint.Nop)
	if _, err := tracedEndpoint(context.Background(), struct{}{}); err != nil {
		t.Fatal(err)
	}

	// tracedEndpoint created a new Span.
	finishedSpans := reporter.Flush()
	if want, have := 1, len(finishedSpans); want != have {
		t.Fatalf("Want %v span(s), found %v", want, have)
	}

	endpointSpan := finishedSpans[0]
	if want, have := "testOp", endpointSpan.Name; want != have {
		t.Fatalf("Want %q, have %q", want, have)
	}
}

func TestTraceClient(t *testing.T) {
	reporter := recorder.NewReporter()
	defer reporter.Close()
	tracer, _ := zipkin.NewTracer(reporter)

	// Initialize the ctx with a parent Span.
	parentSpan := tracer.StartSpan("parent")
	defer parentSpan.Finish()
	ctx := zipkin.NewContext(context.Background(), parentSpan)

	tracedEndpoint := kitzipkin.TraceClient(tracer, "testOp")(endpoint.Nop)
	if _, err := tracedEndpoint(ctx, struct{}{}); err != nil {
		t.Fatal(err)
	}

	// tracedEndpoint created a new Span.
	finishedSpans := reporter.Flush()
	if want, have := 1, len(finishedSpans); want != have {
		t.Fatalf("Want %v span(s), found %v", want, have)
	}

	endpointSpan := finishedSpans[0]
	if want, have := "testOp", endpointSpan.Name; want != have {
		t.Fatalf("Want %q, have %q", want, have)
	}

	parentContext := parentSpan.Context()
	endpointContext := parentSpan.Context()

	// ... and that the parent ID is set appropriately.
	if want, have := parentContext.ID, endpointContext.ID; want != have {
		t.Errorf("Want ParentID %q, have %q", want, have)
	}
}

func TestTraceClientNoContextSpan(t *testing.T) {
	reporter := recorder.NewReporter()
	defer reporter.Close()
	tracer, _ := zipkin.NewTracer(reporter)

	// Empty/background context.
	tracedEndpoint := kitzipkin.TraceClient(tracer, "testOp")(endpoint.Nop)
	if _, err := tracedEndpoint(context.Background(), struct{}{}); err != nil {
		t.Fatal(err)
	}

	// tracedEndpoint created a new Span.
	finishedSpans := reporter.Flush()
	if want, have := 1, len(finishedSpans); want != have {
		t.Fatalf("Want %v span(s), found %v", want, have)
	}

	endpointSpan := finishedSpans[0]
	if want, have := "testOp", endpointSpan.Name; want != have {
		t.Fatalf("Want %q, have %q", want, have)
	}
}
