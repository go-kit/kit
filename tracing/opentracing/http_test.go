package opentracing_test

import (
	"net/http"
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"golang.org/x/net/context"

	kitot "github.com/go-kit/kit/tracing/opentracing"
	kithttp "github.com/go-kit/kit/transport/http"
)

func TestTraceHTTPRequestRoundtrip(t *testing.T) {
	tracer := mocktracer.New()

	// Initialize the ctx with a Span to inject.
	beforeSpan := tracer.StartSpan("to_inject").(*mocktracer.MockSpan)
	defer beforeSpan.Finish()
	beforeSpan.SetBaggageItem("baggage", "check")
	beforeCtx := opentracing.ContextWithSpan(context.Background(), beforeSpan)

	var toHTTPFunc kithttp.RequestFunc = kitot.ToHTTPRequest(tracer, nil)
	req, _ := http.NewRequest("GET", "http://test.biz/url", nil)
	// Call the RequestFunc.
	afterCtx := toHTTPFunc(beforeCtx, req)

	// The Span should not have changed.
	afterSpan := opentracing.SpanFromContext(afterCtx)
	if beforeSpan != afterSpan {
		t.Errorf("Should not swap in a new span")
	}

	// No spans should have finished yet.
	if want, have := 0, len(tracer.FinishedSpans); want != have {
		t.Errorf("Want %v span(s), found %v", want, have)
	}

	// Use FromHTTPRequest to verify that we can join with the trace given a req.
	var fromHTTPFunc kithttp.RequestFunc = kitot.FromHTTPRequest(tracer, "joined", nil)
	joinCtx := fromHTTPFunc(afterCtx, req)
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
