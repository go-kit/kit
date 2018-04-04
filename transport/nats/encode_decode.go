package nats

import (
	"context"
	"github.com/nats-io/go-nats"
)

// DecodeRequestFunc extracts a user-domain request object from a NATS request.
// It's designed to be used in NATS servers, for server-side endpoints. One
// straightforward DecodeRequestFunc could be something that decodes from the
// NATS request message to the concrete request type.
type DecodeRequestFunc func(_ context.Context, msg *nats.Msg) (interface{}, error)

// EncodeRequestFunc encodes the passed request object into the NATS request
// object. It's designed to be used in NATS clients, for client-side endpoints.
// One straightforward EncodeRequestFunc could something that encodes the object
// directly to the NATS request message.
type EncodeRequestFunc func(_ context.Context, msg interface{}) ([]byte, error)

// EncodeResponseFunc encodes the passed response object to the NATS response
// message. It's designed to be used in NATS servers, for server-side endpoints.
// One straightforward EncodeResponseFunc could be something that encodes the
// object directly to the NATS response message.
type EncodeResponseFunc func(_ context.Context, response interface{}) ([]byte, error)

// DecodeResponseFunc extracts a user-domain response object from a NATS
// response object. It's designed to be used in NATS clients, for client-side
// endpoints. One straightforward DecodeResponseFunc could be something that
// decodes from the NATS response message to the concrete response type.
type DecodeResponseFunc func(_ context.Context, msg *nats.Msg) (interface{}, error)
