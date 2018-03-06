package zipkin_test

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	zipkin "github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/propagation/b3"
	"github.com/openzipkin/zipkin-go/reporter/recorder"

	"github.com/go-kit/kit/log"
	kitzipkin "github.com/go-kit/kit/tracing/zipkin"
)

func TestTraceHTTPRequestRoundtrip(t *testing.T) {
	logger := log.NewNopLogger()
	reporter := recorder.NewReporter()
	defer reporter.Close()

	// we disable shared rpc spans so we can test parent-child relationship
	tracer, _ := zipkin.NewTracer(reporter, zipkin.WithSharedSpans(false))

	// Initialize the ctx with a Span to inject.
	beforeSpan := tracer.StartSpan("to_inject")
	beforeCtx := zipkin.NewContext(context.Background(), beforeSpan)

	toHTTPFunc := kitzipkin.ContextToHTTP(tracer, logger)
	req, _ := http.NewRequest("GET", "http://test.biz/path", nil)
	// Call the RequestFunc.
	afterCtx := toHTTPFunc(beforeCtx, req)

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

	// Use HTTPToContext to verify that we can join with the trace given a req.
	fromHTTPFunc := kitzipkin.HTTPToContext(tracer, "joined", logger)
	joinCtx := fromHTTPFunc(afterCtx, req)
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
		t.Error("SpanID should have changed", joined.SpanContext.ID, before.SpanContext.ID)
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

func TestContextToHTTPTags(t *testing.T) {
	reporter := recorder.NewReporter()
	defer reporter.Close()
	tracer, _ := zipkin.NewTracer(reporter)

	span := tracer.StartSpan("to_inject")
	ctx := zipkin.NewContext(context.Background(), span)
	req, _ := http.NewRequest("GET", "http://test.biz/path", nil)

	kitzipkin.ContextToHTTP(tracer, log.NewNopLogger())(ctx, req)

	expectedTags := map[string]string{
		string(zipkin.TagHTTPMethod): "GET",
		string(zipkin.TagHTTPUrl):    "http://test.biz/path",
	}

	span.Finish()

	finishedSpans := reporter.Flush()
	if want, have := 1, len(finishedSpans); want != have {
		t.Fatalf("Want %v span(s), found %v", want, have)
	}

	if !reflect.DeepEqual(expectedTags, finishedSpans[0].Tags) {
		t.Errorf("Want %q, have %q", expectedTags, finishedSpans[0].Tags)
	}
}

func TestHTTPToContextTags(t *testing.T) {
	reporter := recorder.NewReporter()
	defer reporter.Close()
	tracer, _ := zipkin.NewTracer(reporter)

	parentSpan := tracer.StartSpan("to_extract")
	defer parentSpan.Finish()

	req, _ := http.NewRequest("GET", "http://test.biz/path", nil)
	b3.InjectHTTP(req)(parentSpan.Context())

	ctx := kitzipkin.HTTPToContext(tracer, "op", log.NewNopLogger())(context.Background(), req)
	zipkin.SpanFromContext(ctx).Finish()

	childSpan := reporter.Flush()[0]
	expectedTags := map[string]string{
		string(zipkin.TagHTTPMethod): "GET",
		string(zipkin.TagHTTPUrl):    "http://test.biz/path",
	}
	if !reflect.DeepEqual(expectedTags, childSpan.Tags) {
		t.Errorf("Want %q, have %q", expectedTags, childSpan.Tags)
	}
	if want, have := "op", childSpan.Name; want != have {
		t.Errorf("Want %q, have %q", want, have)
	}
}
