package codec

import (
	"io"

	"golang.org/x/net/context"
)

// Codec decodes and encodes requests and responses. Decode takes and returns
// a context because the request or response may be accompanied by information
// that needs to be applied there.
type Codec interface {
	Decode(context.Context, io.Reader, interface{}) (context.Context, error)
	Encode(io.Writer, interface{}) error
}
