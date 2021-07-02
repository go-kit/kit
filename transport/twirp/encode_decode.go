package twirp

import (
	"context"
)

// DecodeRequestFunc extracts a user-domain request object from a Twirp request.
// It's designed to be used in Twirp servers, for server-side endpoints. One
// straightforward DecodeRequestFunc could be something that decodes from the
// Twirp request message to the concrete request type.
type DecodeRequestFunc func(context.Context, interface{}) (request interface{}, err error)

// EncodeRequestFunc encodes the passed request object into the Twirp request
// object. It's designed to be used in Twirp clients, for client-side endpoints.
// One straightforward EncodeRequestFunc could something that encodes the object
// directly to the Twirp request message.
type EncodeRequestFunc func(context.Context, interface{}) (request interface{}, err error)

// EncodeResponseFunc encodes the passed response object to the Twirp response
// message. It's designed to be used in Twirp servers, for server-side endpoints.
// One straightforward EncodeResponseFunc could be something that encodes the
// object directly to the Twirp response message.
type EncodeResponseFunc func(context.Context, interface{}) (response interface{}, err error)

// DecodeResponseFunc extracts a user-domain response object from a Twirp
// response object. It's designed to be used in Twirp clients, for client-side
// endpoints. One straightforward DecodeResponseFunc could be something that
// decodes from the Twirp response message to the concrete response type.
type DecodeResponseFunc func(context.Context, interface{}) (response interface{}, err error)
