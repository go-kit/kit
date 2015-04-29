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

	"github.com/peterbourgon/gokit/server"
	jsoncodec "github.com/peterbourgon/gokit/transport/codec/json"
	httptransport "github.com/peterbourgon/gokit/transport/http"
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

	endpoint := func(_ context.Context, req server.Request) (server.Response, error) {
		r, ok := req.(*myRequest)
		if !ok {
			return nil, fmt.Errorf("not myRequest (%s)", reflect.TypeOf(req))
		}
		return myResponse{transform(r.In)}, nil
	}

	ctx := context.Background()
	requestType := reflect.TypeOf(myRequest{})
	codec := jsoncodec.New()
	binding := httptransport.NewBinding(ctx, requestType, codec, endpoint)
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
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		t.Fatal(err)
	}

	if want, have := transform(n), r.Out; want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}
