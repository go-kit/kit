package ws_test

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-kit/kit/endpoint"
	"github.com/gorilla/websocket"

	"github.com/go-kit/kit/transport/http/ws"
)

func addBody() io.Reader {
	return body(`{"ws": "2.0", "method": "add", "params": [3, 2], "id": 1}`)
}

func body(in string) io.Reader {
	return strings.NewReader(in)
}

func expectErrorCode(t *testing.T, want int, body []byte) {
	var r ws.Response
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

func nopDecoder(context.Context, int, []byte) (interface{}, error) { return struct{}{}, nil }
func nopEncoder(context.Context, interface{}) (int, []byte, error) { return 1, []byte("[]"), nil }

type mockLogger struct {
	Called   bool
	LastArgs []interface{}
}

func (l *mockLogger) Log(keyvals ...interface{}) error {
	l.Called = true
	l.LastArgs = append(l.LastArgs, keyvals)
	return nil
}

// func TestServerBadDecode(t *testing.T) {
// 	ecm := ws.EndpointCodecMap{
// 		"add": ws.EndpointCodec{
// 			Endpoint: endpoint.Nop,
// 			Decode:   func(context.Context, json.RawMessage) (interface{}, error) { return struct{}{}, errors.New("oof") },
// 			Encode:   nopEncoder,
// 		},
// 	}
// 	logger := mockLogger{}
// 	handler := ws.NewServer(ecm, ws.ServerErrorLogger(&logger))
// 	server := httptest.NewServer(handler)
// 	defer server.Close()
// 	resp, _ := http.Post(server.URL, "application/json", addBody())
// 	buf, _ := ioutil.ReadAll(resp.Body)
// 	if want, have := http.StatusOK, resp.StatusCode; want != have {
// 		t.Errorf("want %d, have %d: %s", want, have, buf)
// 	}
// 	expectErrorCode(t, ws.InternalError, buf)
// 	if !logger.Called {
// 		t.Fatal("Expected logger to be called with error. Wasn't.")
// 	}
// }

// func TestServerBadEndpoint(t *testing.T) {
// 	ecm := ws.EndpointCodecMap{
// 		"add": ws.EndpointCodec{
// 			Endpoint: func(context.Context, interface{}) (interface{}, error) { return struct{}{}, errors.New("oof") },
// 			Decode:   nopDecoder,
// 			Encode:   nopEncoder,
// 		},
// 	}
// 	handler := ws.NewServer(ecm)
// 	server := httptest.NewServer(handler)
// 	defer server.Close()
// 	resp, _ := http.Post(server.URL, "application/json", addBody())
// 	if want, have := http.StatusOK, resp.StatusCode; want != have {
// 		t.Errorf("want %d, have %d", want, have)
// 	}
// 	buf, _ := ioutil.ReadAll(resp.Body)
// 	expectErrorCode(t, ws.InternalError, buf)
// }

// func TestServerBadEncode(t *testing.T) {
// 	ecm := ws.EndpointCodecMap{
// 		"add": ws.EndpointCodec{
// 			Endpoint: endpoint.Nop,
// 			Decode:   nopDecoder,
// 			Encode:   func(context.Context, interface{}) (json.RawMessage, error) { return []byte{}, errors.New("oof") },
// 		},
// 	}
// 	handler := ws.NewServer(ecm)
// 	server := httptest.NewServer(handler)
// 	defer server.Close()
// 	resp, _ := http.Post(server.URL, "application/json", addBody())
// 	if want, have := http.StatusOK, resp.StatusCode; want != have {
// 		t.Errorf("want %d, have %d", want, have)
// 	}
// 	buf, _ := ioutil.ReadAll(resp.Body)
// 	expectErrorCode(t, ws.InternalError, buf)
// }

// func TestServerErrorEncoder(t *testing.T) {
// 	errTeapot := errors.New("teapot")
// 	code := func(err error) int {
// 		if err == errTeapot {
// 			return http.StatusTeapot
// 		}
// 		return http.StatusInternalServerError
// 	}
// 	ecm := ws.EndpointCodecMap{
// 		"add": ws.EndpointCodec{
// 			Endpoint: func(context.Context, interface{}) (interface{}, error) { return struct{}{}, errTeapot },
// 			Decode:   nopDecoder,
// 			Encode:   nopEncoder,
// 		},
// 	}
// 	handler := ws.NewServer(
// 		ecm,
// 		ws.ServerErrorEncoder(func(_ context.Context, err error, w http.ResponseWriter) { w.WriteHeader(code(err)) }),
// 	)
// 	server := httptest.NewServer(handler)
// 	defer server.Close()
// 	resp, _ := http.Post(server.URL, "application/json", addBody())
// 	if want, have := http.StatusTeapot, resp.StatusCode; want != have {
// 		t.Errorf("want %d, have %d", want, have)
// 	}
// }

// func TestCanRejectNonPostRequest(t *testing.T) {
// 	ecm := ws.EndpointCodecMap{}
// 	handler := ws.NewServer(ecm)
// 	server := httptest.NewServer(handler)
// 	defer server.Close()
// 	resp, _ := http.Get(server.URL)
// 	if want, have := http.StatusMethodNotAllowed, resp.StatusCode; want != have {
// 		t.Errorf("want %d, have %d", want, have)
// 	}
// }

// func TestCanRejectInvalidJSON(t *testing.T) {
// 	ecm := ws.EndpointCodecMap{}
// 	handler := ws.NewServer(ecm)
// 	server := httptest.NewServer(handler)
// 	defer server.Close()
// 	resp, _ := http.Post(server.URL, "application/json", body("clearlynotjson"))
// 	if want, have := http.StatusOK, resp.StatusCode; want != have {
// 		t.Errorf("want %d, have %d", want, have)
// 	}
// 	buf, _ := ioutil.ReadAll(resp.Body)
// 	expectErrorCode(t, ws.ParseError, buf)
// }

// func TestServerUnregisteredMethod(t *testing.T) {
// 	ecm := ws.EndpointCodecMap{}
// 	handler := ws.NewServer(ecm)
// 	server := httptest.NewServer(handler)
// 	defer server.Close()
// 	resp, _ := http.Post(server.URL, "application/json", addBody())
// 	if want, have := http.StatusOK, resp.StatusCode; want != have {
// 		t.Errorf("want %d, have %d", want, have)
// 	}
// 	buf, _ := ioutil.ReadAll(resp.Body)
// 	expectErrorCode(t, ws.MethodNotFoundError, buf)
// }

func TestServerHappyPath(t *testing.T) {
	step, response := testServer(t)
	step()
	resp := <-response
	defer resp.Body.Close() // nolint
	buf, _ := ioutil.ReadAll(resp.Body)
	if want, have := http.StatusOK, resp.StatusCode; want != have {
		t.Errorf("want %d, have %d (%s)", want, have, buf)
	}
	var r ws.Response
	err := json.Unmarshal(buf, &r)
	if err != nil {
		t.Fatalf("Cant' decode response. err=%s, body=%s", err, buf)
	}
	if r.JSONRPC != ws.Version {
		t.Fatalf("JSONRPC Version: want=%s, got=%s", ws.Version, r.JSONRPC)
	}
	if r.Error != nil {
		t.Fatalf("Unxpected error on response: %s", buf)
	}
}

// func TestMultipleServerBefore(t *testing.T) {
// 	var done = make(chan struct{})
// 	ecm := ws.EndpointCodecMap{
// 		"add": ws.EndpointCodec{
// 			Endpoint: endpoint.Nop,
// 			Decode:   nopDecoder,
// 			Encode:   nopEncoder,
// 		},
// 	}
// 	handler := ws.NewServer(
// 		ecm,
// 		ws.ServerBefore(func(ctx context.Context, r *http.Request) context.Context {
// 			ctx = context.WithValue(ctx, "one", 1)

// 			return ctx
// 		}),
// 		ws.ServerBefore(func(ctx context.Context, r *http.Request) context.Context {
// 			if _, ok := ctx.Value("one").(int); !ok {
// 				t.Error("Value was not set properly when multiple ServerBefores are used")
// 			}

// 			close(done)
// 			return ctx
// 		}),
// 	)
// 	server := httptest.NewServer(handler)
// 	defer server.Close()
// 	http.Post(server.URL, "application/json", addBody()) // nolint

// 	select {
// 	case <-done:
// 	case <-time.After(time.Second):
// 		t.Fatal("timeout waiting for finalizer")
// 	}
// }

// func TestMultipleServerAfter(t *testing.T) {
// 	var done = make(chan struct{})
// 	ecm := ws.EndpointCodecMap{
// 		"add": ws.EndpointCodec{
// 			Endpoint: endpoint.Nop,
// 			Decode:   nopDecoder,
// 			Encode:   nopEncoder,
// 		},
// 	}
// 	handler := ws.NewServer(
// 		ecm,
// 		ws.ServerAfter(func(ctx context.Context, w http.ResponseWriter) context.Context {
// 			ctx = context.WithValue(ctx, "one", 1)

// 			return ctx
// 		}),
// 		ws.ServerAfter(func(ctx context.Context, w http.ResponseWriter) context.Context {
// 			if _, ok := ctx.Value("one").(int); !ok {
// 				t.Error("Value was not set properly when multiple ServerAfters are used")
// 			}

// 			close(done)
// 			return ctx
// 		}),
// 	)
// 	server := httptest.NewServer(handler)
// 	defer server.Close()
// 	http.Post(server.URL, "application/json", addBody()) // nolint

// 	select {
// 	case <-done:
// 	case <-time.After(time.Second):
// 		t.Fatal("timeout waiting for finalizer")
// 	}
// }

// func TestCanFinalize(t *testing.T) {
// 	var done = make(chan struct{})
// 	var finalizerCalled bool
// 	ecm := ws.EndpointCodecMap{
// 		"add": ws.EndpointCodec{
// 			Endpoint: endpoint.Nop,
// 			Decode:   nopDecoder,
// 			Encode:   nopEncoder,
// 		},
// 	}
// 	handler := ws.NewServer(
// 		ecm,
// 		ws.ServerFinalizer(func(ctx context.Context, code int, req *http.Request) {
// 			finalizerCalled = true
// 			close(done)
// 		}),
// 	)
// 	server := httptest.NewServer(handler)
// 	defer server.Close()
// 	http.Post(server.URL, "application/json", addBody()) // nolint

// 	select {
// 	case <-done:
// 	case <-time.After(time.Second):
// 		t.Fatal("timeout waiting for finalizer")
// 	}

// 	if !finalizerCalled {
// 		t.Fatal("Finalizer was not called.")
// 	}
// }

type endpointMapper struct {
	e   endpoint.Endpoint
	dec ws.DecodeRequestFunc
	enc ws.EncodeResponseFunc
}

func (em *endpointMapper) Map([]byte) (endpoint.Endpoint, ws.DecodeRequestFunc, ws.EncodeResponseFunc) {
	return em.e, em.dec, em.enc
}

func testServer(t *testing.T) (step func(), resp <-chan *http.Response) {
	var (
		stepch   = make(chan bool)
		endpoint = func(ctx context.Context, request interface{}) (response interface{}, err error) {
			<-stepch
			return struct{}{}, nil
		}
		response = make(chan *http.Response)
		ecm      = endpointMapper{
			e:   endpoint,
			dec: nopDecoder,
			enc: nopEncoder,
		}
		handler = ws.NewServer(&ecm, websocket.Upgrader{})
	)
	go func() {
		server := httptest.NewServer(handler)
		defer server.Close()
		rb := strings.NewReader(`{"ws": "2.0", "method": "add", "params": [3, 2], "id": 1}`)
		resp, err := http.Post(server.URL, "application/json", rb)
		if err != nil {
			t.Error(err)
			return
		}
		response <- resp
	}()
	return func() { stepch <- true }, response
}
