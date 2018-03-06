package zipkin_test

import (
	"context"
	"testing"

	zipkin "github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/reporter/recorder"
	"google.golang.org/grpc/metadata"

	"github.com/go-kit/kit/log"
	kitzipkin "github.com/go-kit/kit/tracing/zipkin"
)

func TestTraceGRPCRequestRoundtrip(t *testing.T) {
	logger := log.NewNopLogger()
	reporter := recorder.NewReporter()
	defer reporter.Close()

	// we disable shared rpc spans so we can test parent-child relationship
	tracer, _ := zipkin.NewTracer(reporter, zipkin.WithSharedSpans(false))

	// Initialize the ctx with a Span to inject.
	beforeSpan := tracer.StartSpan("to_inject")
	beforeCtx := zipkin.NewContext(context.Background(), beforeSpan)

	toGRPCFunc := kitzipkin.ContextToGRPC(tracer, logger)
	md := metadata.Pairs()
	// Call the RequestFunc.
	afterCtx := toGRPCFunc(beforeCtx, &md)

	// The Span should not have changed.
	afterSpan := zipkin.SpanFromContext(afterCtx)
	if beforeSpan != afterSpan {
		t.Error("Should not swap in a new span")
	}

	// No spans should have finished yet.
	finishedSpans := reporter.Flush()
	if want, have := 0, len(finishedSpans); want != have {
		t.Errorf("Want %v span(s), found %v", want, have)
	}

	// Use GRPCToContext to verify that we can join with the trace given MD.
	fromGRPCFunc := kitzipkin.GRPCToContext(tracer, "joined", logger)
	joinCtx := fromGRPCFunc(afterCtx, md)
	joinedSpan := zipkin.SpanFromContext(joinCtx)

	joinedSpan.Finish()
	beforeSpan.Finish()

	finishedSpans = reporter.Flush()
	if want, have := 2, len(finishedSpans); want != have {
		t.Fatalf("Want %v span(s), found %v", want, have)
	}

	joined := finishedSpans[0]
	before := finishedSpans[1]

	if joined.SpanContext.ID == before.SpanContext.ID {
		t.Error("Span.ID should have changed", joined.SpanContext.ID, before.SpanContext.ID)
	}

	// Check that the parent/child relationship is as expected for the joined span.
	if joined.SpanContext.ParentID == nil {
		t.Fatalf("Want ParentID %q, have nil", before.SpanContext.ID)
	}
	if want, have := before.SpanContext.ID, *joined.SpanContext.ParentID; want != have {
		t.Errorf("Want ParentID %q, have %q", want, have)
	}
	if want, have := "joined", joined.Name; want != have {
		t.Errorf("Want %q, have %q", want, have)
	}
}
