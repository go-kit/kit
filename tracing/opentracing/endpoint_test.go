package opentracing_test

import (
	"context"
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/pkg/errors"

	"github.com/go-kit/kit/endpoint"
	kitot "github.com/go-kit/kit/tracing/opentracing"
)

var (
	endpointErr = errors.New("endpoint error")
	svcErr      = errors.New("svc error")
)

type svcResponse struct {
	err error
}

func (r svcResponse) Failed() error { return r.err }

func passEndpoint(_ context.Context, req interface{}) (interface{}, error) {
	if err, _ := req.(error); err != nil {
		return nil, err
	}
	return req, nil
}

func TestTraceServer(t *testing.T) {
	tracer := mocktracer.New()

	// Initialize the ctx with a nameless Span.
	contextSpan := tracer.StartSpan("").(*mocktracer.MockSpan)
	ctx := opentracing.ContextWithSpan(context.Background(), contextSpan)

	tracedEndpoint := kitot.TraceServer(tracer, "testOp")(endpoint.Nop)
	if _, err := tracedEndpoint(ctx, struct{}{}); err != nil {
		t.Fatal(err)
	}

	finishedSpans := tracer.FinishedSpans()
	if want, have := 1, len(finishedSpans); want != have {
		t.Fatalf("Want %v span(s), found %v", want, have)
	}

	// Test that the op name is updated
	endpointSpan := finishedSpans[0]
	if want, have := "testOp", endpointSpan.OperationName; want != have {
		t.Fatalf("Want %q, have %q", want, have)
	}
	contextContext := contextSpan.Context().(mocktracer.MockSpanContext)
	endpointContext := endpointSpan.Context().(mocktracer.MockSpanContext)
	// ...and that the ID is unmodified.
	if want, have := contextContext.SpanID, endpointContext.SpanID; want != have {
		t.Errorf("Want SpanID %q, have %q", want, have)
	}
}

func TestTraceServerNoContextSpan(t *testing.T) {
	tracer := mocktracer.New()

	// Empty/background context.
	tracedEndpoint := kitot.TraceServer(tracer, "testOp")(endpoint.Nop)
	if _, err := tracedEndpoint(context.Background(), struct{}{}); err != nil {
		t.Fatal(err)
	}

	// tracedEndpoint created a new Span.
	finishedSpans := tracer.FinishedSpans()
	if want, have := 1, len(finishedSpans); want != have {
		t.Fatalf("Want %v span(s), found %v", want, have)
	}

	endpointSpan := finishedSpans[0]
	if want, have := "testOp", endpointSpan.OperationName; want != have {
		t.Fatalf("Want %q, have %q", want, have)
	}
}

func TestTraceServerError(t *testing.T) {
	ctx := context.Background()
	tracer := mocktracer.New()

	tracedEndpoint := kitot.TraceServer(tracer, "testOp")(passEndpoint)

	// first span with an error returned from the endpoint
	tracedEndpoint(ctx, endpointErr)

	// second span with a business error in the response
	tracedEndpoint(ctx, svcResponse{svcErr})

	finishedSpans := tracer.FinishedSpans()
	if want, have := 2, len(finishedSpans); want != have {
		t.Fatalf("Want %v span(s), found %v", want, have)
	}

	spanContainsError(t, finishedSpans[0], endpointErr)
	spanContainsError(t, finishedSpans[1], svcErr)
}

func TestTraceClient(t *testing.T) {
	tracer := mocktracer.New()

	// Initialize the ctx with a parent Span.
	parentSpan := tracer.StartSpan("parent").(*mocktracer.MockSpan)
	defer parentSpan.Finish()
	ctx := opentracing.ContextWithSpan(context.Background(), parentSpan)

	tracedEndpoint := kitot.TraceClient(tracer, "testOp")(endpoint.Nop)
	if _, err := tracedEndpoint(ctx, struct{}{}); err != nil {
		t.Fatal(err)
	}

	// tracedEndpoint created a new Span.
	finishedSpans := tracer.FinishedSpans()
	if want, have := 1, len(finishedSpans); want != have {
		t.Fatalf("Want %v span(s), found %v", want, have)
	}

	endpointSpan := finishedSpans[0]
	if want, have := "testOp", endpointSpan.OperationName; want != have {
		t.Fatalf("Want %q, have %q", want, have)
	}

	parentContext := parentSpan.Context().(mocktracer.MockSpanContext)
	endpointContext := parentSpan.Context().(mocktracer.MockSpanContext)

	// ... and that the parent ID is set appropriately.
	if want, have := parentContext.SpanID, endpointContext.SpanID; want != have {
		t.Errorf("Want ParentID %q, have %q", want, have)
	}
}

func TestTraceClientNoContextSpan(t *testing.T) {
	tracer := mocktracer.New()

	// Empty/background context.
	tracedEndpoint := kitot.TraceClient(tracer, "testOp")(endpoint.Nop)
	if _, err := tracedEndpoint(context.Background(), struct{}{}); err != nil {
		t.Fatal(err)
	}

	// tracedEndpoint created a new Span.
	finishedSpans := tracer.FinishedSpans()
	if want, have := 1, len(finishedSpans); want != have {
		t.Fatalf("Want %v span(s), found %v", want, have)
	}

	endpointSpan := finishedSpans[0]
	if want, have := "testOp", endpointSpan.OperationName; want != have {
		t.Fatalf("Want %q, have %q", want, have)
	}
}

func TestTraceClientError(t *testing.T) {
	ctx := context.Background()
	tracer := mocktracer.New()

	tracedEndpoint := kitot.TraceClient(tracer, "testOp")(passEndpoint)

	// first span with an error returned from the endpoint
	tracedEndpoint(ctx, endpointErr)

	// second span with a business error in the response
	tracedEndpoint(ctx, svcResponse{svcErr})

	finishedSpans := tracer.FinishedSpans()
	if want, have := 2, len(finishedSpans); want != have {
		t.Fatalf("Want %v span(s), found %v", want, have)
	}

	spanContainsError(t, finishedSpans[0], endpointErr)
	spanContainsError(t, finishedSpans[1], svcErr)
}

func spanContainsError(t *testing.T, span *mocktracer.MockSpan, err error) {
	if want, have := true, span.Tag("error"); want != have {
		t.Fatal("Expected span to have error tag")
	}

	if want, have := 1, len(span.Logs()); want != have {
		t.Fatalf("Want %d log entries on span, have %d", want, have)
	}

	log := span.Logs()[0]
	if want, have := 1, len(log.Fields); want != have {
		t.Fatalf("Want %d fields on span log, have %d", want, have)
	}

	if want, have := "error.object", log.Fields[0].Key; want != have {
		t.Fatalf("Want %q key on log entry, have %q", want, have)
	}

	if want, have := err.Error(), log.Fields[0].ValueString; want != have {
		t.Fatalf("Want %q value on log entry, have %q", want, have)
	}
}
