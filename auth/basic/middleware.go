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

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
)

// AuthError represents an authoriation error.
type AuthError struct {
	Realm string
}

// StatusCode is an iimplementation of the StatusCoder interface in go-kit/http.
func (AuthError) StatusCode() int {
	return http.StatusUnauthorized
}

// Error is an implementation of the Error interface.
func (AuthError) Error() string {
	return http.StatusText(http.StatusUnauthorized)
}

// Headers is an implemntation of the Headerer interface in go-kit/http.
func (e AuthError) Headers() http.Header {
	return http.Header{
		"Content-Type":           []string{"text/plain; charset=utf-8"},
		"X-Content-Type-Options": []string{"nosniff"},
		"WWW-Authenticate":       []string{fmt.Sprintf(`Basic realm=%q`, e.Realm)}}
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

// AuthMiddleware returns a Basic Authentication middleware for a particular user and password.
func AuthMiddleware(requiredUser, requiredPassword, realm string) endpoint.Middleware {
	requiredUserBytes := sha256.Sum256([]byte(requiredUser))
	requiredPassBytes := sha256.Sum256([]byte(requiredPassword))

	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			auth := ctx.Value(httptransport.ContextKeyRequestAuthorization).(string)
			givenUser, givenPass, ok := parseBasicAuth(auth)
			if !ok {
				return nil, AuthError{realm}
			}

			// Equalize lengths of supplied and required credentials by hashing them.
			givenUserBytes := sha256.Sum256(givenUser)
			givenPassBytes := sha256.Sum256(givenPass)

			// Compare the supplied credentials to those set in our options.
			if subtle.ConstantTimeCompare(givenUserBytes[:], requiredUserBytes[:]) == 0 ||
				subtle.ConstantTimeCompare(givenPassBytes[:], requiredPassBytes[:]) == 0 {
				return nil, AuthError{realm}
			}

			return next(ctx, request)
		}
	}
}
