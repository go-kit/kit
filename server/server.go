package server

import (
	"errors"
	"io"
	"net/http"

	"golang.org/x/net/context"
)

var (
	// ErrBadCast is an internal error.
	ErrBadCast = errors.New("bad cast")
)

// Request TODO
type Request interface{}

// Response TODO
type Response interface{}

// Service represents a single method.
type Service func(context.Context, Request) (Response, error)

// Codec TODO
type Codec interface {
	Decode(ctx context.Context, src io.Reader) (Request, error)
	Encode(dst io.Writer, resp Response) error
}

// HTTPService TODO
func HTTPService(c Codec, s Service) http.Handler {
	ctx := context.Background()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO if deadline/timeout specified, use a different constructor?
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		// TODO populate with trace ID, etc.

		// Codecs may also populate the context with information.
		req, err := c.Decode(ctx, r.Body)
		r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp, err := s(ctx, req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := c.Encode(w, resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}
