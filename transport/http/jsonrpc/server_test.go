package jsonrpc_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/transport/http/jsonrpc"
)

func addBody() io.Reader {
	return strings.NewReader(`{"jsonrpc": "2.0", "method": "add", "params": [3, 2], "id": 1}`)
}

func expectErrorCode(t *testing.T, want int, body []byte) {
	var r jsonrpc.Response
	err := json.Unmarshal(body, &r)
	if err != nil {
		t.Fatalf("Cant' decode response. err=%s, body=%s", err, body)
	}
	if r.Error == nil {
		t.Fatalf("Expected error on response. Got none: %s", body)
	}
	if have := r.Error.Code; want != have {
		t.Fatalf("Unexpected error code. Want %d, have %d: %s", want, have, body)
	}
}

func TestServerBadDecode(t *testing.T) {
	ecm := jsonrpc.EndpointCodecMap{
		"add": jsonrpc.EndpointCodec{
			Endpoint: func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil },
			Decode:   func(context.Context, json.RawMessage) (interface{}, error) { return struct{}{}, errors.New("oof") },
			Encode:   func(context.Context, interface{}) (json.RawMessage, error) { return []byte{}, nil },
		},
	}
	handler := jsonrpc.NewServer(context.TODO(), ecm)
	server := httptest.NewServer(handler)
	defer server.Close()
	resp, _ := http.Post(server.URL, "application/json", addBody())
	buf, _ := ioutil.ReadAll(resp.Body)
	if want, have := http.StatusOK, resp.StatusCode; want != have {
		t.Errorf("want %d, have %d: %s", want, have, buf)
	}
	expectErrorCode(t, jsonrpc.InternalError, buf)
}

func TestServerBadEndpoint(t *testing.T) {
	ecm := jsonrpc.EndpointCodecMap{
		"add": jsonrpc.EndpointCodec{
			Endpoint: func(context.Context, interface{}) (interface{}, error) { return struct{}{}, errors.New("oof") },
			Decode:   func(context.Context, json.RawMessage) (interface{}, error) { return struct{}{}, nil },
			Encode:   func(context.Context, interface{}) (json.RawMessage, error) { return []byte{}, nil },
		},
	}
	handler := jsonrpc.NewServer(context.TODO(), ecm)
	server := httptest.NewServer(handler)
	defer server.Close()
	resp, _ := http.Post(server.URL, "application/json", addBody())
	if want, have := http.StatusOK, resp.StatusCode; want != have {
		t.Errorf("want %d, have %d", want, have)
	}
	buf, _ := ioutil.ReadAll(resp.Body)
	expectErrorCode(t, jsonrpc.InternalError, buf)
}

func TestServerBadEncode(t *testing.T) {
	ecm := jsonrpc.EndpointCodecMap{
		"add": jsonrpc.EndpointCodec{
			Endpoint: func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil },
			Decode:   func(context.Context, json.RawMessage) (interface{}, error) { return struct{}{}, nil },
			Encode:   func(context.Context, interface{}) (json.RawMessage, error) { return []byte{}, errors.New("oof") },
		},
	}
	handler := jsonrpc.NewServer(context.TODO(), ecm)
	server := httptest.NewServer(handler)
	defer server.Close()
	resp, _ := http.Post(server.URL, "application/json", addBody())
	if want, have := http.StatusOK, resp.StatusCode; want != have {
		t.Errorf("want %d, have %d", want, have)
	}
	buf, _ := ioutil.ReadAll(resp.Body)
	expectErrorCode(t, jsonrpc.InternalError, buf)
}

func TestServerErrorEncoder(t *testing.T) {
	//ecm := jsonrpc.EndpointCodecMap{
	//"add": jsonrpc.EndpointCodec{
	//Endpoint: func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil },
	//Decode:   func(context.Context, json.RawMessage) (interface{}, error) { return struct{}{}, nil },
	//Encode:   func(context.Context, interface{}) (json.RawMessage, error) { return []byte{}, errors.New("oof") },
	//},
	//}
	//handler := jsonrpc.NewServer(context.TODO(), ecm)
	//server := httptest.NewServer(handler)
	//defer server.Close()
	//resp, _ := http.Post(server.URL, "application/json", addBody())
	//if want, have := http.StatusOK, resp.StatusCode; want != have {
	//t.Errorf("want %d, have %d", want, have)
	//}
	//buf, _ := ioutil.ReadAll(resp.Body)
	//expectErrorCode(t, jsonrpc.InternalError, buf)
}
func TestServerUnregisteredMethod(t *testing.T) {
	ecm := jsonrpc.EndpointCodecMap{}
	handler := jsonrpc.NewServer(context.TODO(), ecm)
	server := httptest.NewServer(handler)
	defer server.Close()
	resp, _ := http.Post(server.URL, "application/json", addBody())
	if want, have := http.StatusOK, resp.StatusCode; want != have {
		t.Errorf("want %d, have %d", want, have)
	}
	buf, _ := ioutil.ReadAll(resp.Body)
	expectErrorCode(t, jsonrpc.MethodNotFoundError, buf)
}

func TestServerHappyPath(t *testing.T) {
	step, response := testServer(t)
	step()
	resp := <-response
	defer resp.Body.Close() // nolint
	buf, _ := ioutil.ReadAll(resp.Body)
	if want, have := http.StatusOK, resp.StatusCode; want != have {
		t.Errorf("want %d, have %d (%s)", want, have, buf)
	}
	var r jsonrpc.Response
	err := json.Unmarshal(buf, &r)
	if err != nil {
		t.Fatalf("Cant' decode response. err=%s, body=%s", err, buf)
	}
	if r.Error != nil {
		t.Fatalf("Unxpected error on response: %s", buf)
	}
}

func TestMultipleServerBefore(t *testing.T) {
	var done = make(chan struct{})
	ecm := jsonrpc.EndpointCodecMap{
		"add": jsonrpc.EndpointCodec{
			Endpoint: endpoint.Nop,
			Decode:   func(context.Context, json.RawMessage) (interface{}, error) { return struct{}{}, nil },
			Encode:   func(context.Context, interface{}) (json.RawMessage, error) { return []byte{}, errors.New("oof") },
		},
	}
	handler := jsonrpc.NewServer(
		context.TODO(),
		ecm,
		jsonrpc.ServerBefore(func(ctx context.Context, r *http.Request) context.Context {
			ctx = context.WithValue(ctx, "one", 1)

			return ctx
		}),
		jsonrpc.ServerBefore(func(ctx context.Context, r *http.Request) context.Context {
			if _, ok := ctx.Value("one").(int); !ok {
				t.Error("Value was not set properly when multiple ServerBefores are used")
			}

			close(done)
			return ctx
		}),
	)
	server := httptest.NewServer(handler)
	defer server.Close()
	http.Post(server.URL, "application/json", addBody()) // nolint

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for finalizer")
	}
}

func TestMultipleServerAfter(t *testing.T) {
	var done = make(chan struct{})
	ecm := jsonrpc.EndpointCodecMap{
		"add": jsonrpc.EndpointCodec{
			Endpoint: endpoint.Nop,
			Decode:   func(context.Context, json.RawMessage) (interface{}, error) { return struct{}{}, nil },
			Encode:   func(context.Context, interface{}) (json.RawMessage, error) { return []byte{}, errors.New("oof") },
		},
	}
	handler := jsonrpc.NewServer(
		context.TODO(),
		ecm,
		jsonrpc.ServerAfter(func(ctx context.Context, w http.ResponseWriter) context.Context {
			ctx = context.WithValue(ctx, "one", 1)

			return ctx
		}),
		jsonrpc.ServerAfter(func(ctx context.Context, w http.ResponseWriter) context.Context {
			if _, ok := ctx.Value("one").(int); !ok {
				t.Error("Value was not set properly when multiple ServerAfters are used")
			}

			close(done)
			return ctx
		}),
	)
	server := httptest.NewServer(handler)
	defer server.Close()
	http.Post(server.URL, "application/json", addBody()) // nolint

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for finalizer")
	}
}

func testServer(t *testing.T) (step func(), resp <-chan *http.Response) {
	var (
		stepch   = make(chan bool)
		endpoint = func(ctx context.Context, request interface{}) (response interface{}, err error) {
			<-stepch
			return struct{}{}, nil
		}
		response = make(chan *http.Response)
		ctx      = context.TODO()
		ecm      = jsonrpc.EndpointCodecMap{
			"add": jsonrpc.EndpointCodec{
				Endpoint: endpoint,
				Decode:   func(context.Context, json.RawMessage) (interface{}, error) { return struct{}{}, nil },
				Encode:   func(context.Context, interface{}) (json.RawMessage, error) { return []byte("[]"), nil },
			},
		}
		handler = jsonrpc.NewServer(ctx, ecm)
	)
	go func() {
		server := httptest.NewServer(handler)
		defer server.Close()
		rb := strings.NewReader(`{"jsonrpc": "2.0", "method": "add", "params": [3, 2], "id": 1}`)
		resp, err := http.Post(server.URL, "application/json", rb)
		if err != nil {
			t.Error(err)
			return
		}
		response <- resp
	}()
	return func() { stepch <- true }, response
}
