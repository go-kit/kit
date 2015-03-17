package cors

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Option changes CORS behavior.
type Option func(*options)

// MaxAge sets the Access-Control-Max-Age header for OPTIONS requests whose
// Access-Control-Request-Method is GET. By default, max age is 10 minutes.
func MaxAge(age time.Duration) Option {
	return func(o *options) { o.maxAge = age }
}

// AllowMethods sets Access-Control-Allow-Methods. By default, only the GET
// method is allowed.
func AllowMethods(methods ...string) Option {
	return func(o *options) { o.allowMethods = methods }
}

// AllowHeaders sets Access-Control-Allow-Headers. By default, the headers
// Accept, Accept-Encoding, Authorization, Content-Type, and Origin are
// allowed.
func AllowHeaders(headers ...string) Option {
	return func(o *options) { o.allowHeaders = headers }
}

// AllowOrigin sets Access-Control-Allow-Origin. By default, * is allowed.
func AllowOrigin(origin string) Option {
	return func(o *options) { o.allowOrigin = origin }
}

type options struct {
	maxAge       time.Duration
	allowMethods []string
	allowHeaders []string
	allowOrigin  string
}

// Middleware returns a chainable HTTP handler that applies headers to allow
// CORS related requests.
func Middleware(opts ...Option) func(http.Handler) http.Handler {
	o := &options{
		maxAge:       10 * time.Minute,
		allowMethods: []string{"GET"},
		allowHeaders: []string{"Accept", "Accept-Encoding", "Authorization", "Content-Type", "Origin"},
		allowOrigin:  "*",
	}
	for _, opt := range opts {
		opt(o)
	}

	// via https://github.com/streadway/handy/blob/master/cors/cors.go

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(o.allowMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(o.allowHeaders, ", "))
			w.Header().Set("Access-Control-Allow-Origin", o.allowOrigin)

			switch r.Method {
			case "OPTIONS":
				if r.Header.Get("Access-Control-Request-Method") == "GET" {
					w.Header().Set("Access-Control-Max-Age", strconv.Itoa(int(o.maxAge/time.Second)))
					return
				}
				w.WriteHeader(http.StatusUnauthorized)

			case "HEAD", "GET":
				next.ServeHTTP(w, r)

			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		})
	}
}
