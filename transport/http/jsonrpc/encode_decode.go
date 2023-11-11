package jsonrpc

import (
	"encoding/json"

	"github.com/openmesh/kit/endpoint"

	"context"
)

// Server-Side Codec

// EndpointCodec defines a server Endpoint and its associated codecs
type EndpointCodec[Req, Res any] struct {
	Endpoint endpoint.Endpoint[Req, Res]
	Decode   DecodeRequestFunc[Req]
	Encode   EncodeResponseFunc[Res]
}

// EndpointCodecMap maps the Request.Method to the proper EndpointCodec
type EndpointCodecMap map[string]EndpointCodec[interface{}, interface{}]

// DecodeRequestFunc extracts a user-domain request object from raw JSON
// It's designed to be used in JSON RPC servers, for server-side endpoints.
// One straightforward DecodeRequestFunc could be something that unmarshals
// JSON from the request body to the concrete request type.
type DecodeRequestFunc[Request any] func(context.Context, json.RawMessage) (request Request, err error)

// EncodeResponseFunc encodes the passed response object to a JSON RPC result.
// It's designed to be used in HTTP servers, for server-side endpoints.
// One straightforward EncodeResponseFunc could be something that JSON encodes
// the object directly.
type EncodeResponseFunc[Response any] func(context.Context, Response) (response json.RawMessage, err error)

// Client-Side Codec

// EncodeRequestFunc encodes the given request object to raw JSON.
// It's designed to be used in JSON RPC clients, for client-side
// endpoints. One straightforward EncodeResponseFunc could be something that
// JSON encodes the object directly.
type EncodeRequestFunc[Req any] func(context.Context, Req) (request json.RawMessage, err error)

// DecodeResponseFunc extracts a user-domain response object from an JSON RPC
// response object. It's designed to be used in JSON RPC clients, for
// client-side endpoints. It is the responsibility of this function to decide
// whether any error present in the JSON RPC response should be surfaced to the
// client endpoint.
type DecodeResponseFunc[Res any] func(context.Context, Response) (response Res, err error)
