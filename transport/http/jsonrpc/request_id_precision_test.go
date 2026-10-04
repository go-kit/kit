package jsonrpc_test

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/go-kit/kit/transport/http/jsonrpc"
)

func TestRequestIDNumberPrecision(t *testing.T) {
	for _, input := range []string{
		`0`, `-42`, `9223372036854775807`, `18446744073709551615`,
		`-9223372036854775809`, `9007199254740993e0`, `1.234567890123456789`,
		`1e100`, `1e-100`,
	} {
		t.Run(input, func(t *testing.T) {
			var request jsonrpc.Request
			if err := json.Unmarshal([]byte(`{"id":`+input+`}`), &request); err != nil {
				t.Fatal(err)
			}
			encoded, err := json.Marshal(jsonrpc.Response{ID: request.ID})
			if err != nil {
				t.Fatal(err)
			}
			var response struct{ ID json.RawMessage }
			if err := json.Unmarshal(encoded, &response); err != nil {
				t.Fatal(err)
			}
			if string(response.ID) != input {
				t.Errorf("response ID = %s, want %s", response.ID, input)
			}
			wantInt, intErr := strconv.Atoi(input)
			gotInt, err := request.ID.Int()
			if (err == nil) != (intErr == nil) || (err == nil && gotInt != wantInt) {
				t.Errorf("Int() = %d, %v; want %d, error present %t", gotInt, err, wantInt, intErr != nil)
			}
			wantFloat, floatErr := strconv.ParseFloat(input, 32)
			gotFloat, err := request.ID.Float32()
			if (err == nil) != (floatErr == nil) || (err == nil && gotFloat != float32(wantFloat)) {
				t.Errorf("Float32() = %g, %v; want %g, error present %t", gotFloat, err, wantFloat, floatErr != nil)
			}
			if _, err := request.ID.String(); err == nil {
				t.Error("numeric ID unexpectedly accepted by String()")
			}
		})
	}
}

func TestRequestIDNumberReuse(t *testing.T) {
	var id jsonrpc.RequestID
	for _, input := range []string{`18446744073709551615`, `"text"`, `1e100`, `0`, `"123"`} {
		data := []byte(input)
		if err := json.Unmarshal(data, &id); err != nil {
			t.Fatal(err)
		}
		for i := range data {
			data[i] = 'x'
		}
		encoded, err := json.Marshal(&id)
		if err != nil {
			t.Fatal(err)
		}
		if string(encoded) != input {
			t.Errorf("reused ID = %s, want %s after input buffer changed", encoded, input)
		}
	}
	encoded, err := json.Marshal(&jsonrpc.RequestID{})
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) != "0" {
		t.Errorf("zero value ID = %s, want 0", encoded)
	}
}

func TestRequestIDNumberHTTPResponses(t *testing.T) {
	for _, stage := range []string{"success", "decode", "endpoint", "encode", "missing"} {
		t.Run(stage, func(t *testing.T) {
			failure := errors.New("endpoint codec failed")
			handler := jsonrpc.NewServer(jsonrpc.EndpointCodecMap{
				"echo": {
					Decode: func(context.Context, json.RawMessage) (interface{}, error) {
						if stage == "decode" {
							return nil, failure
						}
						return "ok", nil
					},
					Endpoint: func(_ context.Context, request interface{}) (interface{}, error) {
						if stage == "endpoint" {
							return nil, failure
						}
						return request, nil
					},
					Encode: func(_ context.Context, response interface{}) (json.RawMessage, error) {
						if stage == "encode" {
							return nil, failure
						}
						return json.Marshal(response)
					},
				},
			})
			server := httptest.NewServer(handler)
			defer server.Close()
			method := "echo"
			if stage == "missing" {
				method = "missing"
			}
			for _, id := range []string{`18446744073709551615`, `9007199254740993e0`, `1.234567890123456789`, `1e100`, `0`, `"123"`, `null`} {
				t.Run(id, func(t *testing.T) {
					body := strings.NewReader(`{"jsonrpc":"2.0","method":"` + method + `","params":[],"id":` + id + `}`)
					response, err := http.Post(server.URL, jsonrpc.ContentType, body)
					if err != nil {
						t.Fatal(err)
					}
					defer response.Body.Close()
					data, err := ioutil.ReadAll(response.Body)
					if err != nil {
						t.Fatal(err)
					}
					if response.StatusCode != http.StatusOK {
						t.Fatalf("HTTP status = %d, body = %s", response.StatusCode, data)
					}
					var result struct {
						JSONRPC string
						ID      json.RawMessage
						Result  json.RawMessage
						Error   *jsonrpc.Error
					}
					if err := json.Unmarshal(data, &result); err != nil {
						t.Fatal(err)
					}
					if result.JSONRPC != jsonrpc.Version || string(result.ID) != id {
						t.Errorf("response = %s, want version 2.0 and ID %s", data, id)
					}
					if stage == "success" {
						if result.Error != nil || string(result.Result) != `"ok"` {
							t.Errorf("success response = %s", data)
						}
					} else {
						wantCode := jsonrpc.InternalError
						if stage == "missing" {
							wantCode = jsonrpc.MethodNotFoundError
						}
						if result.Error == nil || result.Error.Code != wantCode || len(result.Result) != 0 {
							t.Errorf("error response = %s, want code %d and no result", data, wantCode)
						}
					}
				})
			}
		})
	}
}
