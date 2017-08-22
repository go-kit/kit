package http

import "context"
import "encoding/json"

import "net/http"

import httptransport "github.com/go-kit/kit/transport/http"

func NewHTTPHandler(endpoints Endpoints) http.Handler {
	m := http.NewServeMux()
	m.Handle("/foo", httptransport.NewServer(endpoints.Foo, DecodeFooRequest, EncodeFooResponse))
	return m
}
func DecodeFooRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req FooRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}
func EncodeFooResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}
