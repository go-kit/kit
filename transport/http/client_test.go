package http_test

import (
	"io"
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
		decode    = func(r io.Reader) (interface{}, error) { return struct{}{}, nil }
		encode    = func(w io.Writer, response interface{}) error { return nil }
		headers   = make(chan string, 1)
		headerKey = "X-Foo"
		headerVal = "abcde"
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers <- r.Header.Get(headerKey)
		w.WriteHeader(http.StatusOK)
	}))

	client := httptransport.Client{
		Method:     "GET",
		URL:        mustParse(server.URL),
		Context:    context.Background(),
		DecodeFunc: decode,
		EncodeFunc: encode,
		Before:     []httptransport.RequestFunc{httptransport.SetRequestHeader(headerKey, headerVal)},
	}

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
