package jwt_test

import (
	"fmt"
	"testing"

	"google.golang.org/grpc/metadata"

	"github.com/briankassouf/kit/auth/jwt"
	"golang.org/x/net/context"
)

func TestToGRPCContext(t *testing.T) {
	md := metadata.MD{}
	md["Authorization"] = []string{fmt.Sprintf("Bearer %s", signedKey)}
	ctx := context.Background()
	reqFunc := jwt.ToGRPCContext()

	ctx = reqFunc(ctx, &md)
	token, ok := ctx.Value("jwtToken").(string)
	if !ok {
		t.Fatal("JWT Token not passed to context correctly")
	}

	if token != signedKey {
		t.Fatalf("JWT tokens did not match: expecting %s got %s", signedKey, token)
	}
}

func TestFromGRPCContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), "jwtToken", signedKey)

	reqFunc := jwt.FromGRPCContext()
	md := metadata.MD{}
	reqFunc(ctx, &md)
	token, ok := md["jwttoken"]
	if !ok {
		t.Fatal("JWT Token not passed to metadata correctly")
	}

	if token[0] != signedKey {
		t.Fatalf("JWT tokens did not match: expecting %s got %s", signedKey, token[0])
	}
}
