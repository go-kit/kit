package server

import (
	"errors"
	"io"
	"net/http"
)

var (
	ErrBadCast = errors.New("bad cast")
)

type Request interface{}
type Response interface{}

type Service func(Request) (Response, error)

type Codec interface {
	Decode(src io.Reader) (Request, error)
	Encode(dst io.Writer, resp Response) error
}

func HTTPServer(c Codec, s Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := c.Decode(r.Body)
		r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp, err := s(req)
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
