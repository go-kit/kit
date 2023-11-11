package basic

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/openmesh/kit/endpoint"
	httptransport "github.com/openmesh/kit/transport/http"
)

// AuthError represents an authorization error.
type AuthError struct {
	Realm string
}

// StatusCode is an implementation of the StatusCoder interface in openmeshsh/http.
func (AuthError) StatusCode() int {
	return http.StatusUnauthorized
}

// Error is an implementation of the Error interface.
func (AuthError) Error() string {
	return http.StatusText(http.StatusUnauthorized)
}

// Headers is an implementation of the Headerer interface in openmesh/http.
func (e AuthError) Headers() http.Header {
	return http.Header{
		"Content-Type":           []string{"text/plain; charset=utf-8"},
		"X-Content-Type-Options": []string{"nosniff"},
		"WWW-Authenticate":       []string{fmt.Sprintf(`Basic realm=%q`, e.Realm)},
	}
}

// parseBasicAuth parses an HTTP Basic Authentication string.
// "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==" returns ([]byte("Aladdin"), []byte("open sesame"), true).
func parseBasicAuth(auth string) (username, password []byte, ok bool) {
	const prefix = "Basic "
	if !strings.HasPrefix(auth, prefix) {
		return
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return
	}

	s := bytes.IndexByte(c, ':')
	if s < 0 {
		return
	}
	return c[:s], c[s+1:], true
}

// Returns a hash of a given slice.
func toHashSlice(s []byte) []byte {
	hash := sha256.Sum256(s)
	return hash[:]
}

// AuthMiddleware returns a Basic Authentication middleware for a particular user and password.
func AuthMiddleware[Request, Response any](requiredUser, requiredPassword, realm string) endpoint.Middleware[Request, Response] {
	requiredUserBytes := toHashSlice([]byte(requiredUser))
	requiredPasswordBytes := toHashSlice([]byte(requiredPassword))

	return func(next endpoint.Endpoint[Request, Response]) endpoint.Endpoint[Request, Response] {
		return func(ctx context.Context, request Request) (Response, error) {
			auth, ok := ctx.Value(httptransport.ContextKeyRequestAuthorization).(string)
			if !ok {
				return *new(Response), AuthError{realm}
			}

			givenUser, givenPassword, ok := parseBasicAuth(auth)
			if !ok {
				return *new(Response), AuthError{realm}
			}

			givenUserBytes := toHashSlice(givenUser)
			givenPasswordBytes := toHashSlice(givenPassword)

			if subtle.ConstantTimeCompare(givenUserBytes, requiredUserBytes) == 0 ||
				subtle.ConstantTimeCompare(givenPasswordBytes, requiredPasswordBytes) == 0 {
				return *new(Response), AuthError{realm}
			}

			return next(ctx, request)
		}
	}
}
