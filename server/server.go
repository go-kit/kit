package server

import (
	"errors"

	"golang.org/x/net/context"
)

// Request is an RPC request.
type Request interface{}

// Response is an RPC response.
type Response interface{}

// Endpoint is the fundamental building block of package server.
// It represents a single RPC method.
type Endpoint func(context.Context, Request) (Response, error)

// ErrBadCast indicates a type error during decoding or encoding.
var ErrBadCast = errors.New("bad cast")

// ErrContextCanceled indicates a controlling context was canceled before the
// request could be served.
var ErrContextCanceled = errors.New("context was canceled")
