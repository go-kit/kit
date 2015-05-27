package json

import (
	"encoding/json"
	"io"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/transport/codec"
)

type jsonCodec struct{}

// New returns a JSON codec. Request and response structures should have
// properly-tagged fields.
func New() codec.Codec { return jsonCodec{} }

func (jsonCodec) Decode(ctx context.Context, r io.Reader, v interface{}) (context.Context, error) {
	return ctx, json.NewDecoder(r).Decode(v)
}

func (jsonCodec) Encode(w io.Writer, v interface{}) error {
	return json.NewEncoder(w).Encode(v)
}
