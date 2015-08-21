package main

import (
	"encoding/json"
	"net/http"

	"golang.org/x/net/context"

	"gopkg.in/kit.v0/addsvc/reqrep"
	"gopkg.in/kit.v0/endpoint"
	httptransport "gopkg.in/kit.v0/transport/http"
)

func makeHTTPBinding(ctx context.Context, e endpoint.Endpoint, before []httptransport.BeforeFunc, after []httptransport.AfterFunc) http.Handler {
	decode := func(r *http.Request) (interface{}, error) {
		var request reqrep.AddRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			return nil, err
		}
		r.Body.Close()
		return request, nil
	}
	encode := func(w http.ResponseWriter, response interface{}) error {
		return json.NewEncoder(w).Encode(response)
	}
	return httptransport.Server{
		Context:    ctx,
		Endpoint:   e,
		DecodeFunc: decode,
		EncodeFunc: encode,
		Before:     before,
		After:      append([]httptransport.AfterFunc{httptransport.SetContentType("application/json; charset=utf-8")}, after...),
	}
}
