package addsvc

// This file provides server-side bindings for the HTTP transport.
// It utilizes the transport/http.Server.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	stdopentracing "github.com/opentracing/opentracing-go"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/opentracing"
	httptransport "github.com/go-kit/kit/transport/http"
)

// MakeHTTPHandler returns a handler that makes a set of endpoints available
// on predefined paths.
func MakeHTTPHandler(endpoints Endpoints, tracer stdopentracing.Tracer, logger log.Logger) http.Handler {
	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(errorEncoder),
		httptransport.ServerErrorLogger(logger),
	}
	m := http.NewServeMux()
	m.Handle("/sum", httptransport.NewServer(
		endpoints.SumEndpoint,
		DecodeHTTPSumRequest,
		EncodeHTTPGenericResponse,
		append(options, httptransport.ServerBefore(opentracing.FromHTTPRequest(tracer, "Sum", logger)))...,
	))
	m.Handle("/concat", httptransport.NewServer(
		endpoints.ConcatEndpoint,
		DecodeHTTPConcatRequest,
		EncodeHTTPGenericResponse,
		append(options, httptransport.ServerBefore(opentracing.FromHTTPRequest(tracer, "Concat", logger)))...,
	))
	return m
}

func errorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	code := http.StatusInternalServerError
	msg := err.Error()

	switch err {
	case ErrTwoZeroes, ErrMaxSizeExceeded, ErrIntOverflow:
		code = http.StatusBadRequest
	}

	w.WriteHeader(code)
	json.NewEncoder(w).Encode(errorWrapper{Error: msg})
}

func errorDecoder(r *http.Response) error {
	var w errorWrapper
	if err := json.NewDecoder(r.Body).Decode(&w); err != nil {
		return err
	}
	return errors.New(w.Error)
}

type errorWrapper struct {
	Error string `json:"error"`
}

// DecodeHTTPSumRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded sum request from the HTTP request body. Primarily useful in a
// server.
func DecodeHTTPSumRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req sumRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

// DecodeHTTPConcatRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded concat request from the HTTP request body. Primarily useful in a
// server.
func DecodeHTTPConcatRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req concatRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

// DecodeHTTPSumResponse is a transport/http.DecodeResponseFunc that decodes a
// JSON-encoded sum response from the HTTP response body. If the response has a
// non-200 status code, we will interpret that as an error and attempt to decode
// the specific error message from the response body. Primarily useful in a
// client.
func DecodeHTTPSumResponse(_ context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode != http.StatusOK {
		return nil, errorDecoder(r)
	}
	var resp sumResponse
	err := json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}

// DecodeHTTPConcatResponse is a transport/http.DecodeResponseFunc that decodes
// a JSON-encoded concat response from the HTTP response body. If the response
// has a non-200 status code, we will interpret that as an error and attempt to
// decode the specific error message from the response body. Primarily useful in
// a client.
func DecodeHTTPConcatResponse(_ context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode != http.StatusOK {
		return nil, errorDecoder(r)
	}
	var resp concatResponse
	err := json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}

// EncodeHTTPGenericRequest is a transport/http.EncodeRequestFunc that
// JSON-encodes any request to the request body. Primarily useful in a client.
func EncodeHTTPGenericRequest(_ context.Context, r *http.Request, request interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return err
	}
	r.Body = ioutil.NopCloser(&buf)
	return nil
}

// EncodeHTTPGenericResponse is a transport/http.EncodeResponseFunc that encodes
// the response as JSON to the response writer. Primarily useful in a server.
func EncodeHTTPGenericResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}
