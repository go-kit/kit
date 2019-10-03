package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/cmd/kitgen/testdata/foo/default/endpoints"
	httptransport "github.com/go-kit/kit/transport/http"
)

func NewHTTPHandler(endpoints endpoints.Endpoints) http.Handler {
	m := http.NewServeMux()
	m.Handle("/bar", httptransport.NewServer(endpoints.Bar, DecodeBarRequest, EncodeBarResponse))
	return m
}
func DecodeBarRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoints.BarRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}
func EncodeBarResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}
