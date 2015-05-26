package http_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"golang.org/x/net/context"

	jsoncodec "github.com/go-kit/kit/transport/codec/json"
	httptransport "github.com/go-kit/kit/transport/http"
)

func TestBinding(t *testing.T) {
	type myRequest struct {
		In int `json:"in"`
	}

	type myResponse struct {
		Out int `json:"out"`
	}

	transform := func(i int) int {
		return 3 * i // doesn't matter, just do something
	}

	endpoint := func(_ context.Context, request interface{}) (interface{}, error) {
		r, ok := request.(*myRequest)
		if !ok {
			return nil, fmt.Errorf("not myRequest (%s)", reflect.TypeOf(request))
		}
		return myResponse{transform(r.In)}, nil
	}

	ctx := context.Background()
	makeRequest := func() interface{} { return &myRequest{} }
	codec := jsoncodec.New()
	binding := httptransport.NewBinding(ctx, makeRequest, codec, endpoint)
	server := httptest.NewServer(binding)
	defer server.Close()

	n := 123
	requestBody, err := json.Marshal(myRequest{n})
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var r myResponse
	if _, err := codec.Decode(ctx, resp.Body, &r); err != nil {
		t.Fatal(err)
	}

	if want, have := transform(n), r.Out; want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}
