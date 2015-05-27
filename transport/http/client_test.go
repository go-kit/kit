package http_test

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"golang.org/x/net/context"

	jsoncodec "github.com/go-kit/kit/transport/codec/json"
	httptransport "github.com/go-kit/kit/transport/http"
)

func TestClient(t *testing.T) {
	type myResponse struct {
		V int `json:"v"`
	}
	const v = 123
	codec := jsoncodec.New()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		codec.Encode(w, myResponse{v})
	}))
	defer server.Close()
	makeResponse := func() interface{} { return &myResponse{} }
	client := httptransport.NewClient(server.URL, codec, makeResponse)
	resp, err := client(context.Background(), struct{}{})
	if err != nil {
		t.Fatal(err)
	}
	response, ok := resp.(*myResponse)
	if !ok {
		t.Fatalf("not myResponse (%s)", reflect.TypeOf(response))
	}
	if want, have := v, response.V; want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}
