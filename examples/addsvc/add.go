package main

import (
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/examples/addsvc/reqrep"
	"github.com/go-kit/kit/log"
)

// Add is the abstract definition of what this service does. It could easily
// be an interface type with multiple methods, in which case each method would
// be an endpoint.
type Add func(context.Context, int64, int64) int64

// pureAdd implements Add with no dependencies.
func pureAdd(_ context.Context, a, b int64) int64 { return a + b }

// proxyAdd returns an implementation of Add that invokes a remote Add
// service.
func proxyAdd(remote endpoint.Endpoint, logger log.Logger) Add {
	return func(ctx context.Context, a, b int64) int64 {
		resp, err := remote(ctx, reqrep.AddRequest{A: a, B: b})
		if err != nil {
			logger.Log("err", err)
			return 0
		}
		addResp, ok := resp.(reqrep.AddResponse)
		if !ok {
			logger.Log("err", endpoint.ErrBadCast)
			return 0
		}
		return addResp.V
	}
}
