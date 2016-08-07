package auth

import (
	"errors"
	"fmt"

	"google.golang.org/grpc/metadata"

	"golang.org/x/net/context"

	"github.com/briankassouf/kit/transport/grpc"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/endpoint"
)

// Create a new JWT token generating middleware, specifying signing method and the claims
// you would like it to contain. Particulary useful for clients.
func NewJWTSigner(key string, method jwt.SigningMethod, claims jwt.Claims) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			token := jwt.NewWithClaims(method, claims)

			// Sign and get the complete encoded token as a string using the secret
			tokenString, err := token.SignedString([]byte(key))
			if err != nil {
				return nil, err
			}
			md := metadata.Pairs("jwtToken", tokenString)
			ctx = metadata.NewContext(ctx, md)

			return next(ctx, request)
		}
	}
}

// Create a new JWT token parsing middleware, specifying a jwt.Keyfunc interface and the
// signing method. Adds the resulting claims to endpoint context or returns error on invalid
// token. Particualry useful for servers.
func NewJWTParser(keyFunc jwt.Keyfunc, method jwt.SigningMethod) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			// tokenString is stored in the context from the transport handlers
			tokenString, ok := ctx.Value("jwtToken").(string)
			if !ok {
				return nil, errors.New("Token up for parsing was not passed through the context")
			}

			// Parse takes the token string and a function for looking up the key. The latter is especially
			// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
			// head of the token to identify which key to use, but the parsed token (head and claims) is provided
			// to the callback, providing flexibility.
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Don't forget to validate the alg is what you expect:
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
				}
				return keyFunc(token)
			})
			if err != nil {
				return nil, err
			}

			if !token.Valid {
				return nil, errors.New("Could not parse JWT Token")
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				ctx = context.WithValue(ctx, "jwtClaims", claims)
			}

			return next(ctx, request)
		}
	}
}

func NewGRPCServerRequestFunc() grpc.RequestFunc {
	return func(ctx context.Context, md *metadata.MD) context.Context {
		token, ok := (*md)["jwttoken"]
		if ok {
			ctx = context.WithValue(ctx, "jwtToken", token[0])
		}

		return ctx
	}
}

func NewGRPCClientRequestFunc() grpc.RequestFunc {
	return func(ctx context.Context, md *metadata.MD) context.Context {
		md1, ok := metadata.FromContext(ctx)
		if !ok {
			return ctx
		}

		token, ok := md1["jwttoken"]
		if ok {
			(*md)["jwttoken"] = token
		}

		return ctx
	}
}
