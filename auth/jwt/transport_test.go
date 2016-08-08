package jwt_test

import (
	"fmt"
	"testing"

	"google.golang.org/grpc/metadata"

	"github.com/go-kit/kit/auth/jwt"
	"golang.org/x/net/context"
)

func TestToGRPCContext(t *testing.T) {
	md := metadata.MD{}
	reqFunc := jwt.ToGRPCContext()

	// No Authorization header is passed
	ctx := reqFunc(context.Background(), &md)
	token := ctx.Value(jwt.JWTTokenContextKey)
	if token != nil {
		t.Fatal("Context should not contain a JWT Token")
	}

	// Invalid Authorization header is passed
	md["authorization"] = []string{fmt.Sprintf("%s", signedKey)}
	ctx = reqFunc(context.Background(), &md)
	token = ctx.Value(jwt.JWTTokenContextKey)
	if token != nil {
		t.Fatal("Context should not contain a JWT Token")
	}

	// Authorization header is correct
	md["authorization"] = []string{fmt.Sprintf("Bearer %s", signedKey)}
	ctx = reqFunc(context.Background(), &md)
	token, ok := ctx.Value(jwt.JWTTokenContextKey).(string)
	if !ok {
		t.Fatal("JWT Token not passed to context correctly")
	}

	if token != signedKey {
		t.Fatalf("JWT tokens did not match: expecting %s got %s", signedKey, token)
	}
}

func TestFromGRPCContext(t *testing.T) {
	reqFunc := jwt.FromGRPCContext()

	// No JWT Token is passed in the context
	ctx := context.Background()
	md := metadata.MD{}
	reqFunc(ctx, &md)

	_, ok := md["authorization"]
	if ok {
		t.Fatal("authorization key should not exist in metadata")
	}

	// Correct JWT Token is passed in the context
	ctx = metadata.NewContext(context.Background(), metadata.MD{jwt.JWTTokenContextKey: []string{signedKey}})
	md = metadata.MD{}
	reqFunc(ctx, &md)

	token, ok := md["authorization"]
	if !ok {
		t.Fatal("JWT Token not passed to metadata correctly")
	}

	if token[0] != generateAuthHeaderFromToken(signedKey) {
		t.Fatalf("JWT tokens did not match: expecting %s got %s", signedKey, token[0])
	}
}

func generateAuthHeaderFromToken(token string) string {
	return fmt.Sprintf("Bearer %s", token)
}
