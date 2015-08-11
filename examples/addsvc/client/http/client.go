package http

import (
	"bytes"
	"encoding/json"
	"net/http"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/examples/addsvc/reqrep"
	httptransport "github.com/go-kit/kit/transport/http"
)

// NewClient takes a URL that should point to an instance of an addsvc. It
// returns an endpoint that makes a request to that URL.
func NewClient(method, url string, before ...httptransport.RequestFunc) endpoint.Endpoint {
	return func(ctx0 context.Context, request interface{}) (interface{}, error) {
		var (
			ctx, cancel = context.WithCancel(ctx0)
			errs        = make(chan error, 1)
			responses   = make(chan interface{}, 1)
		)
		defer cancel()
		go func() {
			var buf bytes.Buffer
			if err := json.NewEncoder(&buf).Encode(request); err != nil {
				errs <- err
				return
			}
			req, err := http.NewRequest(method, url, &buf)
			if err != nil {
				errs <- err
				return
			}
			for _, f := range before {
				ctx = f(ctx, req)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				errs <- err
				return
			}
			defer resp.Body.Close()
			var response reqrep.AddResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				errs <- err
				return
			}
			responses <- response
		}()
		select {
		case <-ctx.Done():
			return nil, context.DeadlineExceeded
		case err := <-errs:
			return nil, err
		case response := <-responses:
			return response, nil
		}
	}
}
