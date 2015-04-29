package main

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/peterbourgon/gokit/metrics"
	"github.com/peterbourgon/gokit/server"
)

// jsonCodec implements transport/codec, decoding and encoding requests and
// responses respectively as JSON. It requires that the concrete request and
// response types support JSON de/serialization.
//
// This type is mostly boiler-plate; in theory, it could be generated.
type jsonCodec struct{}

func (jsonCodec) Decode(ctx context.Context, r io.Reader) (server.Request, context.Context, error) {
	var req request
	err := json.NewDecoder(r).Decode(&req)
	return &req, ctx, err
}

func (jsonCodec) Encode(w io.Writer, resp server.Response) error {
	return json.NewEncoder(w).Encode(resp)
}

// The HTTP binding exists in the HTTP transport package, because it uses the
// codec to deserialize and serialize requests and responses, and therefore
// doesn't need to have access to the concrete request and response types.

func httpInstrument(requests metrics.Counter, duration metrics.Histogram) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			defer func(begin time.Time) { duration.Observe(time.Since(begin).Nanoseconds()) }(time.Now())
			next.ServeHTTP(w, r)
		})
	}
}
