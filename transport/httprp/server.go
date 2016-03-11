package httprp

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"golang.org/x/net/context"
)

// RequestFunc may take information from an HTTP request and put it into a
// request context. In Servers, BeforeFuncs are executed prior to invoking the
// endpoint. In Clients, BeforeFuncs are executed after creating the request
// but prior to invoking the HTTP client.
type RequestFunc func(context.Context, *http.Request) context.Context

// Server wraps an endpoint and implements http.Handler.
type Server struct {
	ctx          context.Context
	proxy        http.Handler
	before       []RequestFunc
	errorEncoder func(w http.ResponseWriter, err error)
}

// NewServer constructs a new server, which implements http.Server and wraps
// the provided endpoint.
func NewServer(
	ctx context.Context,
	target *url.URL,
	options ...ServerOption,
) *Server {
	s := &Server{
		ctx:   ctx,
		proxy: httputil.NewSingleHostReverseProxy(target),
	}
	for _, option := range options {
		option(s)
	}
	return s
}

// ServerOption sets an optional parameter for servers.
type ServerOption func(*Server)

// ServerBefore functions are executed on the HTTP request object before the
// request is decoded.
func ServerBefore(before ...RequestFunc) ServerOption {
	return func(s *Server) { s.before = before }
}

// ServeHTTP implements http.Handler.
func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(s.ctx)
	defer cancel()

	for _, f := range s.before {
		ctx = f(ctx, r)
	}

	s.proxy.ServeHTTP(w, r)
}
