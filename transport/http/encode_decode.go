package http

import "net/http"

// DecodeFunc converts an HTTP request (transport-domain) to a user request
// (business-domain). One straightforward DecodeFunc could be something that
// JSON-decodes the request body to a concrete request type.
type DecodeFunc func(*http.Request) (interface{}, error)

// EncodeFunc converts a user response (business-domain) to an HTTP response
// (transport-domain) by encoding the interface to the response writer.
type EncodeFunc func(http.ResponseWriter, interface{}) error
