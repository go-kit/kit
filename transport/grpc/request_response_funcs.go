package grpc

import (
	"context"
	"encoding/base64"
	"strings"

	"google.golang.org/grpc/metadata"
)

const (
	binHdrSuffix = "-bin"
)

// RequestFunc may take information from a gRPC request and put it into a
// request context. In Servers, RequestFuncs are executed prior to invoking the
// endpoint. In Clients, RequestFuncs are executed after creating the request
// but prior to invoking the gRPC client.
type RequestFunc func(context.Context, *metadata.MD) context.Context

// ServerResponseFunc may take information from a request context and use it to
// manipulate the gRPC metadata header. ResponseFuncs are only executed in
// servers, after invoking the endpoint but prior to writing a response.
type ServerResponseFunc func(context.Context, *metadata.MD)

// ClientResponseFunc may take information from a gRPC metadata header and/or
// trailer and make the responses available for consumption. ClientResponseFuncs
// are only executed in clients, after a request has been made, but prior to it
// being decoded.
type ClientResponseFunc func(ctx context.Context, header *metadata.MD, trailer *metadata.MD) context.Context

// SetResponseHeader returns a ResponseFunc that sets the specified metadata
// key-value pair.
func SetResponseHeader(key, val string) ServerResponseFunc {
	return func(_ context.Context, md *metadata.MD) {
		key, val := EncodeKeyValue(key, val)
		(*md)[key] = append((*md)[key], val)
	}
}

// SetRequestHeader returns a RequestFunc that sets the specified metadata
// key-value pair.
func SetRequestHeader(key, val string) RequestFunc {
	return func(ctx context.Context, md *metadata.MD) context.Context {
		key, val := EncodeKeyValue(key, val)
		(*md)[key] = append((*md)[key], val)
		return ctx
	}
}

// EncodeKeyValue sanitizes a key-value pair for use in gRPC metadata headers.
func EncodeKeyValue(key, val string) (string, string) {
	key = strings.ToLower(key)
	if strings.HasSuffix(key, binHdrSuffix) {
		v := base64.StdEncoding.EncodeToString([]byte(val))
		val = string(v)
	}
	return key, val
}
