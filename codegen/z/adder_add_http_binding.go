// Do not edit! Generated by gokit-generate
package z

import (
	"encoding/json"
	"net/http"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
)

func makeAdderAddHTTPBinding(ctx context.Context, e endpoint.Endpoint, before []httptransport.BeforeFunc, after []httptransport.AfterFunc) http.Handler {
	decode := func(r *http.Request) (interface{}, error) {
		defer r.Body.Close()
		var request AdderAddRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			return nil, err
		}
		return request, nil
	}
	encode := func(w http.ResponseWriter, response interface{}) error {
		return json.NewEncoder(w).Encode(response)
	}
	return httptransport.Server{
		Context:	ctx,
		Endpoint:	e,
		DecodeFunc:	decode,
		EncodeFunc:	encode,
		Before:		before,
		After:		append([]httptransport.AfterFunc{httptransport.SetContentType("application/json; charset=utf-8")}, after...),
	}
}
