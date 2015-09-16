package http_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"golang.org/x/net/context"

	httptransport "github.com/go-kit/kit/transport/http"
)

func TestHTTPClient(t *testing.T) {
	var (
		encode    = func(*http.Request, interface{}) error { return nil }
		decode    = func(*http.Response) (interface{}, error) { return struct{}{}, nil }
		headers   = make(chan string, 1)
		headerKey = "X-Foo"
		headerVal = "abcde"
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers <- r.Header.Get(headerKey)
		w.WriteHeader(http.StatusOK)
	}))

	client := httptransport.NewClient(
		"GET",
		mustParse(server.URL),
		encode,
		decode,
		httptransport.SetClientBefore(httptransport.SetRequestHeader(headerKey, headerVal)),
	)

	_, err := client.Endpoint()(context.Background(), struct{}{})
	if err != nil {
		t.Fatal(err)
	}

	var have string
	select {
	case have = <-headers:
	case <-time.After(time.Millisecond):
		t.Fatalf("timeout waiting for %s", headerKey)
	}
	if want := headerVal; want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}

func mustParse(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}
