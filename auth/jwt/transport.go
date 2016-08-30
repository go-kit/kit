package jwt

import (
	"fmt"
	stdhttp "net/http"
	"strings"

	"github.com/go-kit/kit/transport/grpc"
	"github.com/go-kit/kit/transport/http"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

const (
	BEARER        string = "bearer"
	BEARER_FORMAT string = "Bearer %s"
)

// moves JWT token from request header to context
// particularly useful for servers
func ToHTTPContext() http.RequestFunc {
	return func(ctx context.Context, r *stdhttp.Request) context.Context {
		token, ok := extractTokenFromAuthHeader(r.Header.Get("Authorization"))
		if !ok {
			return ctx
		}

		return context.WithValue(ctx, JWTTokenContextKey, token)
	}
}

// moves JWT token from context to request header
// particularly useful for clients
func FromHTTPContext() http.RequestFunc {
	return func(ctx context.Context, r *stdhttp.Request) context.Context {
		token, ok := ctx.Value(JWTTokenContextKey).(string)
		if ok {
			r.Header.Add("Authorization", generateAuthHeaderFromToken(token))
		}
		return ctx
	}
}

// moves JWT token from grpc metadata to context
// particularly userful for servers
func ToGRPCContext() grpc.RequestFunc {
	return func(ctx context.Context, md *metadata.MD) context.Context {
		// capital "Key" is illegal in HTTP/2.
		authHeader, ok := (*md)["authorization"]
		if !ok {
			return ctx
		}

		token, ok := extractTokenFromAuthHeader(authHeader[0])
		if ok {
			ctx = context.WithValue(ctx, JWTTokenContextKey, token)
		}

		return ctx
	}
}

// moves JWT token from context to grpc metadata
// particularly useful for clients
func FromGRPCContext() grpc.RequestFunc {
	return func(ctx context.Context, md *metadata.MD) context.Context {
		token, ok := ctx.Value(JWTTokenContextKey).(string)
		if ok {
			// capital "Key" is illegal in HTTP/2.
			(*md)["authorization"] = []string{generateAuthHeaderFromToken(token)}
		}

		return ctx
	}
}

// extractTokenFromAuthHeader returns the token from the value of the Authorzation header
func extractTokenFromAuthHeader(val string) (token string, ok bool) {
	authHeaderParts := strings.Split(val, " ")
	if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != BEARER {
		return "", false
	}

	return authHeaderParts[1], true
}

func generateAuthHeaderFromToken(token string) string {
	return fmt.Sprintf(BEARER_FORMAT, token)
}
