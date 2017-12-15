package jsonrpc

import (
	"encoding/json"

	"github.com/go-kit/kit/endpoint"

	"context"
)

// Server-Side Codec

// EndpointCodec defines a server Endpoint and its associated codecs
type EndpointCodec struct {
	Endpoint endpoint.Endpoint
	Decode   DecodeRequestFunc
	Encode   EncodeResponseFunc
}

// EndpointCodecMap maps the Request.Method to the proper EndpointCodec
type EndpointCodecMap map[string]EndpointCodec

// DecodeRequestFunc extracts a user-domain request object from an raw JSON
// It's designed to be used in HTTP servers, for server-side endpoints.
// One straightforward DecodeRequestFunc could be something that unmarshals
// JSON from the request body to the concrete request type.
type DecodeRequestFunc func(context.Context, json.RawMessage) (request interface{}, err error)

// EncodeResponseFunc encodes the passed response object to a JSON RPC response.
// It's designed to be used in HTTP servers, for server-side endpoints.
// One straightforward EncodeResponseFunc could be something that JSON encodes
// the object directly.
type EncodeResponseFunc func(context.Context, interface{}) (response json.RawMessage, err error)

// Client-Side Codec

// EncodeRequestFunc encodes the passed request object to raw JSON.
// It's designed to be used in JSON RPC clients, for client-side
// endpoints. One straightforward EncodeResponseFunc could be something that
// JSON encodes the object directly.
type EncodeRequestFunc func(context.Context, interface{}) (request json.RawMessage, err error)

// DecodeResponseFunc extracts a user-domain response object from an HTTP
// request object. It's designed to be used in JSON RPC clients, for
// client-side endpoints. One straightforward DecodeRequestFunc could be
// something that JSON decodes from the request body to the concrete
// response type.
type DecodeResponseFunc func(context.Context, json.RawMessage) (response interface{}, err error)
