package opentracing_test

import (
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"

	kitot "github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/transport/grpc"
)

func TestTraceGRPCRequestRoundtrip(t *testing.T) {
	tracer := mocktracer.New()

	// Initialize the ctx with a Span to inject.
	beforeSpan := tracer.StartSpan("to_inject").(*mocktracer.MockSpan)
	defer beforeSpan.Finish()
	beforeSpan.SetBaggageItem("baggage", "check")
	beforeCtx := opentracing.ContextWithSpan(context.Background(), beforeSpan)

	var toGRPCFunc grpc.RequestFunc = kitot.ToGRPCRequest(tracer, nil)
	md := metadata.Pairs()
	// Call the RequestFunc.
	afterCtx := toGRPCFunc(beforeCtx, &md)

	// The Span should not have changed.
	afterSpan := opentracing.SpanFromContext(afterCtx)
	if beforeSpan != afterSpan {
		t.Errorf("Should not swap in a new span")
	}

	// No spans should have finished yet.
	if want, have := 0, len(tracer.FinishedSpans); want != have {
		t.Errorf("Want %v span(s), found %v", want, have)
	}

	// Use FromGRPCRequest to verify that we can join with the trace given MD.
	var fromGRPCFunc grpc.RequestFunc = kitot.FromGRPCRequest(tracer, "joined", nil)
	joinCtx := fromGRPCFunc(afterCtx, &md)
	joinedSpan := opentracing.SpanFromContext(joinCtx).(*mocktracer.MockSpan)

	if joinedSpan.SpanID == beforeSpan.SpanID {
		t.Error("SpanID should have changed", joinedSpan.SpanID, beforeSpan.SpanID)
	}

	// Check that the parent/child relationship is as expected for the joined span.
	if want, have := beforeSpan.SpanID, joinedSpan.ParentID; want != have {
		t.Errorf("Want ParentID %q, have %q", want, have)
	}
	if want, have := "joined", joinedSpan.OperationName; want != have {
		t.Errorf("Want %q, have %q", want, have)
	}
	if want, have := "check", joinedSpan.BaggageItem("baggage"); want != have {
		t.Errorf("Want %q, have %q", want, have)
	}
}
