package jsonrpc_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-kit/kit/transport/http/jsonrpc"
)

type TestResponse struct {
	Body   io.ReadCloser
	String string
}

func TestCanCallBeforeFunc(t *testing.T) {
	called := false
	u, _ := url.Parse("http://senseye.io/jsonrpc")
	sut := jsonrpc.NewClient(
		u,
		"add",
		nopEncoder,
		nopDecoder,
		jsonrpc.ClientBefore(func(ctx context.Context, req *http.Request) context.Context {
			called = true
			return ctx
		}),
	)

	sut.Endpoint()(context.TODO(), "foo")

	if !called {
		t.Fatal("Expected client before func to be called. Wasn't.")
	}
}

//func TestCanCallAfterFunc(t *testing.T) {
//called := false
//u, _ := url.Parse("http://senseye.io/jsonrpc")
//sut := jsonrpc.NewClient(
//u,
//"add",
//nopEncoder,
//nopDecoder,
//jsonrpc.ClientAfter(func(ctx context.Context, req *http.Response) context.Context {
//called = true
//return ctx
//}),
//)

//_, err := sut.Endpoint()(context.TODO(), "foo")
//if err != nil {
//t.Fatal(err)
//}

//if !called {
//t.Fatal("Expected client after func to be called. Wasn't.")
//}
//}

func TestClientHappyPath(t *testing.T) {
	var (
		testbody = `{"jsonrpc":"2.0", "result":5}`
		encode   = func(_ context.Context, req interface{}) (json.RawMessage, error) {
			return json.Marshal(req)
		}
		decode = func(ctx context.Context, res json.RawMessage) (interface{}, error) {
			if ac := ctx.Value("afterCalled"); ac == nil {
				t.Fatal("after not called")
			}
			var result int
			err := json.Unmarshal(res, &result)
			if err != nil {
				return nil, err
			}
			return result, nil
		}
		afterFunc = func(ctx context.Context, r *http.Response) context.Context {
			return context.WithValue(ctx, "afterCalled", true)
		}
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testbody))
	}))

	sut := jsonrpc.NewClient(
		mustParse(server.URL),
		"add",
		encode,
		decode,
		jsonrpc.ClientAfter(afterFunc),
	)

	result, err := sut.Endpoint()(context.Background(), struct{}{})
	if err != nil {
		t.Fatal(err)
	}
	ri, ok := result.(int)
	if !ok {
		t.Fatalf("result is not int: (%T)%+v", result, result)
	}
	if ri != 5 {
		t.Fatalf("want=5, got=%d", ri)
	}
}

//func TestClientFinalizer(t *testing.T) {
//var (
//headerKey    = "X-Henlo-Lizer"
//headerVal    = "Helllo you stinky lizard"
//responseBody = "go eat a fly ugly\n"
//done         = make(chan struct{})
//encode       = func(context.Context, *http.Request, interface{}) error { return nil }
//decode       = func(_ context.Context, r *http.Response) (interface{}, error) {
//return TestResponse{r.Body, ""}, nil
//}
//)

//server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//w.Header().Set(headerKey, headerVal)
//w.Write([]byte(responseBody))
//}))
//defer server.Close()

//client := httptransport.NewClient(
//"GET",
//mustParse(server.URL),
//encode,
//decode,
//httptransport.ClientFinalizer(func(ctx context.Context, err error) {
//responseHeader := ctx.Value(httptransport.ContextKeyResponseHeaders).(http.Header)
//if want, have := headerVal, responseHeader.Get(headerKey); want != have {
//t.Errorf("%s: want %q, have %q", headerKey, want, have)
//}

//responseSize := ctx.Value(httptransport.ContextKeyResponseSize).(int64)
//if want, have := int64(len(responseBody)), responseSize; want != have {
//t.Errorf("response size: want %d, have %d", want, have)
//}

//close(done)
//}),
//)

//_, err := client.Endpoint()(context.Background(), struct{}{})
//if err != nil {
//t.Fatal(err)
//}

//select {
//case <-done:
//case <-time.After(time.Second):
//t.Fatal("timeout waiting for finalizer")
//}
//}

//func TestEncodeJSONRequest(t *testing.T) {
//var header http.Header
//var body string

//server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//b, err := ioutil.ReadAll(r.Body)
//if err != nil && err != io.EOF {
//t.Fatal(err)
//}
//header = r.Header
//body = string(b)
//}))

//defer server.Close()

//serverURL, err := url.Parse(server.URL)

//if err != nil {
//t.Fatal(err)
//}

//client := httptransport.NewClient(
//"POST",
//serverURL,
//httptransport.EncodeJSONRequest,
//func(context.Context, *http.Response) (interface{}, error) { return nil, nil },
//).Endpoint()

//for _, test := range []struct {
//value interface{}
//body  string
//}{
//{nil, "null\n"},
//{12, "12\n"},
//{1.2, "1.2\n"},
//{true, "true\n"},
//{"test", "\"test\"\n"},
//{enhancedRequest{Foo: "foo"}, "{\"foo\":\"foo\"}\n"},
//} {
//if _, err := client(context.Background(), test.value); err != nil {
//t.Error(err)
//continue
//}

//if body != test.body {
//t.Errorf("%v: actual %#v, expected %#v", test.value, body, test.body)
//}
//}

//if _, err := client(context.Background(), enhancedRequest{Foo: "foo"}); err != nil {
//t.Fatal(err)
//}

//if _, ok := header["X-Edward"]; !ok {
//t.Fatalf("X-Edward value: actual %v, expected %v", nil, []string{"Snowden"})
//}

//if v := header.Get("X-Edward"); v != "Snowden" {
//t.Errorf("X-Edward string: actual %v, expected %v", v, "Snowden")
//}
//}

func mustParse(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}
