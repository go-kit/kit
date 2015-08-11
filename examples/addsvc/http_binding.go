package main

import (
	"encoding/json"
	"io"
	"net/http"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/examples/addsvc/reqrep"
	httptransport "github.com/go-kit/kit/transport/http"
)

func makeHTTPBinding(ctx context.Context, e endpoint.Endpoint, before []httptransport.RequestFunc, after []httptransport.ResponseFunc) http.Handler {
	decode := func(r io.Reader) (interface{}, error) {
		var request reqrep.AddRequest
		if err := json.NewDecoder(r).Decode(&request); err != nil {
			return nil, err
		}
		return request, nil
	}
	encode := func(w io.Writer, response interface{}) error {
		return json.NewEncoder(w).Encode(response)
	}
	return httptransport.Server{
		Context:    ctx,
		Endpoint:   e,
		DecodeFunc: decode,
		EncodeFunc: encode,
		Before:     before,
		After:      append([]httptransport.ResponseFunc{httptransport.SetContentType("application/json; charset=utf-8")}, after...),
	}
}
