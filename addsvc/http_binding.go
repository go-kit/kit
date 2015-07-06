package main

import (
	"encoding/json"
	"net/http"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/addsvc/reqrep"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
)

type httpBinding struct {
	context.Context
	endpoint.Endpoint
	Before []httptransport.BeforeFunc
	After  []httptransport.AfterFunc
}

func (b httpBinding) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		var request reqrep.AddRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			errcodes <- errcode{err, http.StatusBadRequest}
			return
		}
		r.Body.Close()
		r, err := b.Endpoint(ctx, request)
		if err != nil {
			errcodes <- errcode{err, http.StatusInternalServerError}
			return
		}
		response, ok := r.(reqrep.AddResponse)
		if !ok {
			errcodes <- errcode{endpoint.ErrBadCast, http.StatusInternalServerError}
			return
		}
		for _, f := range b.After {
			f(ctx, w)
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
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
