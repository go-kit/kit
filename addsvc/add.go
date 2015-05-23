package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/zipkin"
)

// Add is the abstract definition of what this service does. It could easily
// be an interface type with multiple methods. Each method would be an
// endpoint.
type Add func(context.Context, int64, int64) int64

func pureAdd(_ context.Context, a, b int64) int64 { time.Sleep(34 * time.Millisecond); return a + b }

func addViaHTTP(addr string, newSpan zipkin.NewSpanFunc, c zipkin.Collector) Add {
	// TODO make & use addsvc HTTP client

	if !strings.HasPrefix(addr, "http") {
		addr = "http://" + addr
	}
	u, err := url.Parse(addr)
	if err != nil {
		panic(err)
	}
	u.Path = "/add"

	return func(ctx context.Context, a, b int64) int64 {
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(map[string]interface{}{"a": a, "b": b}); err != nil {
			log.DefaultLogger.Log("err", err)
			return 0
		}

		req, err := http.NewRequest("GET", u.String(), &buf)
		if err != nil {
			log.DefaultLogger.Log("err", err)
			return 0
		}

		span := zipkin.NewChildSpan(ctx, newSpan)
		defer c.Collect(span)
		span.Annotate(zipkin.ClientSend)
		zipkin.SetRequestHeaders(req.Header, zipkin.NewChildSpan(ctx, newSpan))
		resp, err := http.DefaultClient.Do(req)
		span.Annotate(zipkin.ClientReceive)
		if err != nil {
			log.DefaultLogger.Log("err", err)
			return 0
		}
		defer resp.Body.Close()

		var response struct {
			V int64 `json:"v"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			log.DefaultLogger.Log("err", err)
			return 0
		}

		return response.V
	}
}
