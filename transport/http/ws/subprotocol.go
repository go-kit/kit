package ws

import (
	"context"
	"io"
)

// Server-Side Codec

// SubprotocolCodec defines a WebSocket Subprotocol to use for decoding and encoding data
type SubprotocolCodec struct {
	Decode DecodeSubprotocolFunc
	Encode EncodeSubprotocolFunc
}

// SubprotocolCodecMap maps the WebSocket Sec-WebSocket-Protocol initial handshake header to a SubprotocolCodec
type SubprotocolCodecMap map[string]SubprotocolCodec

// DecodeSubprotocolFunc extracts the incoming WebSocket request into a string that maps
// to the EndpointsCodecMap and remaining request data into an io.Reader.
type DecodeSubprotocolFunc func(context.Context, io.Reader) (string, io.Reader, error)

// EncodeSubprotocolFunc reads from the io.Reader which contains the results from EncodeResponseFunc
// and writes it to the WebSocket io.Writer to return to the connected client.
type EncodeSubprotocolFunc func(context.Context, string, io.Writer, io.Reader) error
