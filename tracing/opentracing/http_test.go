package opentracing_test

import (
	"net/http"
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"golang.org/x/net/context"

	"github.com/go-kit/kit/log"
	kitot "github.com/go-kit/kit/tracing/opentracing"
)

func TestTraceHTTPRequestRoundtrip(t *testing.T) {
	logger := log.NewNopLogger()
	tracer := mocktracer.New()

	// Initialize the ctx with a Span to inject.
	beforeSpan := tracer.StartSpan("to_inject").(*mocktracer.MockSpan)
	defer beforeSpan.Finish()
	beforeSpan.Context().SetBaggageItem("baggage", "check")
	beforeCtx := opentracing.ContextWithSpan(context.Background(), beforeSpan)

	toHTTPFunc := kitot.ToHTTPRequest(tracer, logger)
	req, _ := http.NewRequest("GET", "http://test.biz/url", nil)
	// Call the RequestFunc.
	afterCtx := toHTTPFunc(beforeCtx, req)

	// The Span should not have changed.
	afterSpan := opentracing.SpanFromContext(afterCtx)
	if beforeSpan != afterSpan {
		t.Errorf("Should not swap in a new span")
	}

	// No spans should have finished yet.
	finishedSpans := tracer.FinishedSpans()
	if want, have := 0, len(finishedSpans); want != have {
		t.Errorf("Want %v span(s), found %v", want, have)
	}

	// Use FromHTTPRequest to verify that we can join with the trace given a req.
	fromHTTPFunc := kitot.FromHTTPRequest(tracer, "joined", logger)
	joinCtx := fromHTTPFunc(afterCtx, req)
	joinedSpan := opentracing.SpanFromContext(joinCtx).(*mocktracer.MockSpan)

	joinedContext := joinedSpan.Context().(*mocktracer.MockSpanContext)
	beforeContext := beforeSpan.Context().(*mocktracer.MockSpanContext)

	if joinedContext.SpanID == beforeContext.SpanID {
		t.Error("SpanID should have changed", joinedContext.SpanID, beforeContext.SpanID)
	}

	// Check that the parent/child relationship is as expected for the joined span.
	if want, have := beforeContext.SpanID, joinedSpan.ParentID; want != have {
		t.Errorf("Want ParentID %q, have %q", want, have)
	}
	if want, have := "joined", joinedSpan.OperationName; want != have {
		t.Errorf("Want %q, have %q", want, have)
	}
	if want, have := "check", joinedSpan.Context().BaggageItem("baggage"); want != have {
		t.Errorf("Want %q, have %q", want, have)
	}
}
