package codec

import (
	"io"

	"golang.org/x/net/context"

	"github.com/peterbourgon/gokit/server"
)

// Codec defines how to decode and encode requests and responses. Decode takes
// a context because the request may be accompanied by information that needs
// to be applied there.
type Codec interface {
	Decode(context.Context, io.Reader) (server.Request, error)
	Encode(io.Writer, server.Response) error
}
