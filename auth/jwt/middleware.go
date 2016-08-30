package jwt

import (
	"errors"

	"golang.org/x/net/context"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/endpoint"
)

const (
	// JWTTokenContextKey holds the key used to store a JWT Token in the context
	JWTTokenContextKey = "JWTToken"
	// JWTClaimsContxtKey holds the key used to store the JWT Claims in the context
	JWTClaimsContextKey = "JWTClaims"
)

var (
	ErrTokenContextMissing     = errors.New("Token up for parsing was not passed through the context")
	ErrTokenInvalid            = errors.New("JWT Token was invalid")
	ErrUnexpectedSigningMethod = errors.New("Unexpected signing method")
	ErrKIDNotFound             = errors.New("Key ID was not found in key set")
	ErrNoKIDHeader             = errors.New("Token doesn't have 'kid' header")
)

type Claims map[string]interface{}

type KeySet map[string]struct {
	Method jwt.SigningMethod
	Key    []byte
}

// Create a new JWT token generating middleware, specifying signing method and the claims
// you would like it to contain. Particularly useful for clients.
func NewSigner(kid string, keys KeySet, claims Claims) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			key, ok := keys[kid]
			if !ok {
				return nil, ErrKIDNotFound
			}

			token := jwt.NewWithClaims(key.Method, jwt.MapClaims(claims))
			token.Header["kid"] = kid
			// Sign and get the complete encoded token as a string using the secret
			tokenString, err := token.SignedString(key.Key)
			if err != nil {
				return nil, err
			}
			ctx = context.WithValue(ctx, JWTTokenContextKey, tokenString)

			return next(ctx, request)
		}
	}
}

// Create a new JWT token parsing middleware, specifying a jwt.Keyfunc interface and the
// signing method. Adds the resulting claims to endpoint context or returns error on invalid
// token. Particularly useful for servers.
func NewParser(keys KeySet) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			// tokenString is stored in the context from the transport handlers
			tokenString, ok := ctx.Value(JWTTokenContextKey).(string)
			if !ok {
				return nil, ErrTokenContextMissing
			}

			// Parse takes the token string and a function for looking up the key. The latter is especially
			// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
			// head of the token to identify which key to use, but the parsed token (head and claims) is provided
			// to the callback, providing flexibility.
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				kid, ok := token.Header["kid"]
				if !ok {
					return nil, ErrNoKIDHeader
				}

				key, ok := keys[kid.(string)]
				if !ok {
					return nil, ErrKIDNotFound
				}

				// Don't forget to validate the alg is what you expect:
				if token.Method != key.Method {
					return nil, ErrUnexpectedSigningMethod
				}

				return key.Key, nil
			})
			if err != nil {
				if e, ok := err.(*jwt.ValidationError); ok && e.Inner != nil {
					return nil, e.Inner
				}

				return nil, err
			}

			if !token.Valid {
				return nil, ErrTokenInvalid
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				ctx = context.WithValue(ctx, JWTClaimsContextKey, Claims(claims))
			}

			return next(ctx, request)
		}
	}
}
