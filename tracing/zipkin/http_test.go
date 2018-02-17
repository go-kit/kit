package zipkin_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/model"
	"github.com/openzipkin/zipkin-go/reporter/recorder"

	zipkinkit "github.com/go-kit/kit/tracing/zipkin"
	kithttp "github.com/go-kit/kit/transport/http"
)

const (
	testName     = "test"
	testBody     = "test_body"
	testTagKey   = "test_key"
	testTagValue = "test_value"
)

func TestHttpClientTracePropagatesParentSpan(t *testing.T) {
	rec := recorder.NewReporter()
	defer rec.Close()

	tr, _ := zipkin.NewTracer(rec)

	rURL, _ := url.Parse("http://test.com")

	clientTracer := zipkinkit.HTTPClientTrace(tr)
	ep := kithttp.NewClient(
		"GET",
		rURL,
		func(ctx context.Context, r *http.Request, i interface{}) error {
			return nil
		},
		func(ctx context.Context, r *http.Response) (response interface{}, err error) {
			return nil, nil
		},
		clientTracer,
	).Endpoint()

	parentSpan := tr.StartSpan("test")

	ctx := zipkin.NewContext(context.Background(), parentSpan)

	_, err := ep(ctx, nil)
	if err != nil {
		t.Fatalf("unwanted error: %s", err.Error())
	}

	spans := rec.Flush()
	if want, have := 1, len(spans); want != have {
		t.Fatalf("incorrect number of spans, wanted %d, got %d", want, have)
	}

	span := spans[0]
	if span.SpanContext.ParentID == nil {
		t.Fatalf("incorrect parent ID, got nil")
	}

	if want, have := parentSpan.Context().ID, *span.SpanContext.ParentID; want != have {
		t.Fatalf("incorrect parent ID, wanted %s, got %s", want, have)
	}
}

func TestHTTPClientTraceAddsExpectedTags(t *testing.T) {
	dataProvider := []struct {
		ResponseStatusCode int
		ErrorTagValue      string
	}{
		{http.StatusOK, ""},
		{http.StatusForbidden, fmt.Sprint(http.StatusForbidden)},
	}

	for _, data := range dataProvider {
		testHTTPClientTraceCase(t, data.ResponseStatusCode, data.ErrorTagValue)
	}
}

func testHTTPClientTraceCase(t *testing.T, responseStatusCode int, errTagValue string) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(responseStatusCode)
		w.Write([]byte(testBody))
	}))
	defer ts.Close()

	rec := recorder.NewReporter()
	defer rec.Close()

	tr, err := zipkin.NewTracer(rec)
	if err != nil {
		t.Errorf("Unwanted error: %s", err.Error())
	}

	rMethod := "GET"
	rURL, _ := url.Parse(ts.URL)

	clientTracer := zipkinkit.HTTPClientTrace(
		tr,
		zipkinkit.Name(testName),
		zipkinkit.Tags(map[string]string{testTagKey: testTagValue}),
	)

	ep := kithttp.NewClient(
		rMethod,
		rURL,
		func(ctx context.Context, r *http.Request, i interface{}) error {
			return nil
		},
		func(ctx context.Context, r *http.Response) (response interface{}, err error) {
			return nil, nil
		},
		clientTracer,
	).Endpoint()

	_, err = ep(context.Background(), nil)
	if err != nil {
		t.Fatalf("unwanted error: %s", err.Error())
	}

	spans := rec.Flush()
	if want, have := 1, len(spans); want != have {
		t.Fatalf("incorrect number of spans, wanted %d, got %d", want, have)
	}

	span := spans[0]
	if span.SpanContext.ParentID != nil {
		t.Fatalf("incorrect parentID, wanted nil, got %s", span.SpanContext.ParentID)
	}

	if want, have := testName, span.Name; want != have {
		t.Fatalf("incorrect span name, wanted %s, got %s", want, have)
	}

	if want, have := model.Client, span.Kind; want != have {
		t.Fatalf("incorrect span kind, wanted %s, got %s", want, have)
	}

	tags := map[string]string{
		testTagKey:                         testTagValue,
		string(zipkin.TagHTTPStatusCode):   fmt.Sprint(responseStatusCode),
		string(zipkin.TagHTTPMethod):       rMethod,
		string(zipkin.TagHTTPUrl):          rURL.String(),
		string(zipkin.TagHTTPResponseSize): fmt.Sprint(len(testBody)),
	}

	if errTagValue != "" {
		tags[string(zipkin.TagError)] = fmt.Sprint(errTagValue)
	}

	if !reflect.DeepEqual(span.Tags, tags) {
		t.Fatalf("invalid tags set, wanted %+v, got %+v", tags, span.Tags)
	}
}
