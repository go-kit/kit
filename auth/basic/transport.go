package basic

import (
	"context"
	"fmt"
	stdhttp "net/http"
	"strings"

	"google.golang.org/grpc/metadata"

	"github.com/go-kit/kit/transport/grpc"
	"github.com/go-kit/kit/transport/http"
)

type contextKey string

const (
	// BasicTokenContextKey holds the key used to store a Basic Token in the
	// context.
	BasicTokenContextKey contextKey = "BasicToken"
)

const (
	basic       string = "basic"
	basicFormat string = "Basic %s"
)

// HTTPToContext moves a basic token from request header to context. Particularly
// useful for servers.
func HTTPToContext() http.RequestFunc {
	return func(ctx context.Context, r *stdhttp.Request) context.Context {
		token, ok := extractTokenFromAuthHeader(r.Header.Get("Authorization"))
		if !ok {
			return ctx
		}

		return context.WithValue(ctx, BasicTokenContextKey, token)
	}
}

// ContextToHTTP moves a basic token from context to request header. Particularly
// useful for clients.
func ContextToHTTP() http.RequestFunc {
	return func(ctx context.Context, r *stdhttp.Request) context.Context {
		token, ok := ctx.Value(BasicTokenContextKey).(string)
		if ok {
			r.Header.Add("Authorization", generateAuthHeaderFromToken(token))
		}
		return ctx
	}
}

// GRPCToContext moves a basic token from grpc metadata to context. Particularly
// userful for servers.
func GRPCToContext() grpc.ServerRequestFunc {
	return func(ctx context.Context, md metadata.MD) context.Context {
		// capital "Key" is illegal in HTTP/2.
		authHeader, ok := md["authorization"]
		if !ok {
			return ctx
		}

		token, ok := extractTokenFromAuthHeader(authHeader[0])
		if ok {
			ctx = context.WithValue(ctx, BasicTokenContextKey, token)
		}

		return ctx
	}
}

// ContextToGRPC moves a basic token from context to grpc metadata. Particularly
// useful for clients.
func ContextToGRPC() grpc.ClientRequestFunc {
	return func(ctx context.Context, md *metadata.MD) context.Context {
		token, ok := ctx.Value(BasicTokenContextKey).(string)
		if ok {
			// capital "Key" is illegal in HTTP/2.
			(*md)["authorization"] = []string{generateAuthHeaderFromToken(token)}
		}

		return ctx
	}
}

func extractTokenFromAuthHeader(val string) (token string, ok bool) {
	authHeaderParts := strings.Split(val, " ")
	if len(authHeaderParts) != 2 || !strings.EqualFold(authHeaderParts[0], basic) {
		return "", false
	}

	return authHeaderParts[1], true
}

func generateAuthHeaderFromToken(token string) string {
	return fmt.Sprintf(basicFormat, token)
}
