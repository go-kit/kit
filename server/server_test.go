package server_test

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/peterbourgon/gokit/server"
	"golang.org/x/net/context"
)

func TestHTTPEndpoint(t *testing.T) {
	ctx := context.Background()
	codec := jsonCodec{}
	endpoint := echoEndpoint
	server := httptest.NewServer(server.HTTPEndpoint(ctx, codec, endpoint))
	defer server.Close()

	value := 42
	buf, err := json.Marshal(echoRequest{value})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.Post(server.URL, "application/json", bytes.NewReader(buf))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if want, have := http.StatusOK, resp.StatusCode; want != have {
		buf, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("want HTTP %d, have %d (%s)", want, have, buf)
	}

	var r echoResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		t.Fatal(err)
	}

	if want, have := value, r.Value; want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

type (
	jsonCodec    struct{}
	echoRequest  struct{ Value int }
	echoResponse struct{ Value int }
)

func (jsonCodec) Encode(w io.Writer, resp server.Response) error {
	return json.NewEncoder(w).Encode(resp)
}

func (jsonCodec) Decode(_ context.Context, r io.Reader) (server.Request, error) {
	var req echoRequest
	err := json.NewDecoder(r).Decode(&req)
	return req, err
}

var echoEndpoint = func(_ context.Context, req server.Request) (server.Response, error) {
	testReq, ok := req.(echoRequest)
	if !ok {
		return nil, server.ErrBadCast
	}
	return echoResponse{Value: testReq.Value}, nil
}
