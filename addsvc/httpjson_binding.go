package main

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/peterbourgon/gokit/metrics"
	"github.com/peterbourgon/gokit/server"
	"github.com/peterbourgon/gokit/server/zipkin"
	"github.com/peterbourgon/gokit/transport/codec"
)

// jsonCodec decodes and encodes requests and responses respectively as JSON.
// It requires that the (package main) request and response structs support
// JSON de/serialization.
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

// A binding wraps an Endpoint so that it's usable by a transport. httpBinding
// makes an Endpoint usable over HTTP. It combines a parent context, a codec,
// and an endpoint to expose. It implements http.Handler by decoding a request
// from the HTTP request body, and encoding a response to the response writer.
type httpBinding struct {
	context.Context        // parent context
	codec.Codec            // how to decode requests and encode responses
	contentType     string // what we report as the response ContentType
	server.Endpoint        // the endpoint being bound
}

func (b httpBinding) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Perform HTTP-specific context amendments.
	b.Context = zipkin.GetHeaders(b.Context, r.Header)

	// Decode request.
	req, ctx, err := b.Codec.Decode(b.Context, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	b.Context = ctx

	// Execute RPC.
	resp, err := b.Endpoint(b.Context, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Encode response.
	w.Header().Set("Content-Type", b.contentType)
	if err := b.Codec.Encode(w, resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func httpInstrument(requests metrics.Counter, duration metrics.Histogram, next http.Handler) http.Handler {
	return httpInstrumented{requests, duration, next}
}

type httpInstrumented struct {
	requests metrics.Counter
	duration metrics.Histogram
	next     http.Handler
}

func (i httpInstrumented) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	i.requests.Add(1)
	defer func(begin time.Time) { i.duration.Observe(time.Since(begin).Nanoseconds()) }(time.Now())
	i.next.ServeHTTP(w, r)
}
