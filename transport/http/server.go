package http

import (
	"net/http"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
)

// Server wraps an endpoint and implements http.Handler.
type Server struct {
	context.Context
	endpoint.Endpoint
	DecodeFunc
	EncodeFunc
	Before []BeforeFunc
	After  []AfterFunc
}

// ServeHTTP implements http.Handler.
func (b Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	type errcode struct {
		error
		int
	}
	var (
		ctx, cancel = context.WithCancel(b.Context)
		errcodes    = make(chan errcode, 1)
		done        = make(chan struct{}, 1)
	)
	defer cancel()
	go func() {
		for _, f := range b.Before {
			ctx = f(ctx, r)
		}
		request, err := b.DecodeFunc(r)
		if err != nil {
			errcodes <- errcode{err, http.StatusBadRequest}
			return
		}
		response, err := b.Endpoint(ctx, request)
		if err != nil {
			errcodes <- errcode{err, http.StatusInternalServerError}
			return
		}
		for _, f := range b.After {
			f(ctx, w)
		}
		if err := b.EncodeFunc(w, response); err != nil {
			errcodes <- errcode{err, http.StatusInternalServerError}
			return
		}
		close(done)
	}()
	select {
	case <-ctx.Done():
		http.Error(w, context.DeadlineExceeded.Error(), http.StatusInternalServerError)
	case errcode := <-errcodes:
		http.Error(w, errcode.error.Error(), errcode.int)
	case <-done:
		return
	}
}
