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

func ToHTTPContext() http.RequestFunc {
	return func(ctx context.Context, r *stdhttp.Request) context.Context {
		token, ok := extractTokenFromAuthHeader(r.Header.Get("Authorization"))
		if !ok {
			return ctx
		}

		return context.WithValue(ctx, "jwtToken", token)
	}
}

func ToGRPCContext() grpc.RequestFunc {
	return func(ctx context.Context, md *metadata.MD) context.Context {
		authHeader, ok := (*md)["Authorization"]
		if !ok {
			return ctx
		}

		token, ok := extractTokenFromAuthHeader(authHeader[0])
		if ok {
			ctx = context.WithValue(ctx, "jwtToken", token)
		}

		return ctx
	}
}

func FromGRPCContext() grpc.RequestFunc {
	return func(ctx context.Context, md *metadata.MD) context.Context {
		md1, ok := metadata.FromContext(ctx)
		if !ok {
			return ctx
		}

		token, ok := md1["jwttoken"]
		if ok {
			(*md)["Authorization"] = []string{generateAuthHeaderFromToken(token[0])}
		}

		return ctx
	}
}

// extractTokenFromAuthHeader returns the token from the value of the Authorzation header
func extractTokenFromAuthHeader(val string) (token string, ok bool) {
	if len(val) < 8 || !strings.EqualFold(val[0:7], "BEARER ") {
		return "", false
	}

	return val[7:], true
}

func generateAuthHeaderFromToken(token string) string {
	return fmt.Sprintf("Bearer %s", token)
}
