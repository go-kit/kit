package http

import "io"

// DecodeFunc converts a serialized request (server) or response (client) to a
// user version of the same. One straightforward DecodeFunc could be something
// that JSON-decodes the reader to a concrete type.
type DecodeFunc func(io.Reader) (interface{}, error)

// EncodeFunc converts a user response (server) or request (client) to a
// serialized version of the same, by encoding the interface to the writer.
type EncodeFunc func(io.Writer, interface{}) error
