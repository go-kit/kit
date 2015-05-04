package json

import (
	"encoding/json"
	"io"

	"golang.org/x/net/context"

	"github.com/peterbourgon/gokit/server"
	"github.com/peterbourgon/gokit/transport/codec"
)

type jsonCodec struct{}

// New returns a JSON codec. Request and response structures should have
// properly-tagged fields.
func New() codec.Codec { return jsonCodec{} }

func (jsonCodec) Decode(ctx context.Context, r io.Reader, req server.Request) (context.Context, error) {
	return ctx, json.NewDecoder(r).Decode(req)
}

func (jsonCodec) Encode(w io.Writer, resp server.Response) error {
	return json.NewEncoder(w).Encode(resp)
}
