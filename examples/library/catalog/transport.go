package catalog

import (
"context"
"encoding/json"
"errors"
"net/http"
"github.com/gorilla/mux"

kitlog "github.com/go-kit/kit/log"
"github.com/go-kit/kit/transport"
kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-kit/kit/examples/library/books"
)

// MakeHandler returns a handler for the booking service.
func MakeHandler(bs Service, logger kitlog.Logger) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		kithttp.ServerErrorEncoder(encodeError),
	}

	bookHandler := kithttp.NewServer(
		makeBookEndpoint(bs),
		decodeBookRequest,
		encodeResponse,
		opts...,
	)

	getBookHandler := kithttp.NewServer(
		makeGetBookEndpoint(bs),
		decodeGetBookRequest,
		encodeResponse,
		opts...,
	)

	r := mux.NewRouter()

	r.Handle("/library/v1/book", bookHandler).Methods("POST")
	r.Handle("/library/v1/book", getBookHandler).Methods("GET")
	return r
}

var errBadRoute = errors.New("bad route")

func decodeBookRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var body struct {
		Name string `json:"Name,omitempty"`
		Author string `json:"Author,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}
	return bookRequest{
		Name: body.Name,
		Author: body.Author,
	}, nil
}

func decodeGetBookRequest(ctx context.Context, r *http.Request) (interface{}, error) {

	name := r.FormValue("name")
	return bookRequest{
		Name: name,

	}, nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

type errorer interface {
	error() error
}

// encode errors from business-logic
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch err {
	case books.ErrUnknown:
		w.WriteHeader(http.StatusNotFound)
	case ErrInvalidArgument:
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

