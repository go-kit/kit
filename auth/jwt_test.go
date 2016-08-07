package auth_test

import (
	"testing"

	"google.golang.org/grpc/metadata"

	jwt "github.com/dgrijalva/jwt-go"

	"github.com/go-kit/kit/auth"
	"golang.org/x/net/context"
)

var (
	key       = "test_signing_key"
	method    = jwt.SigningMethodHS256
	claims    = jwt.MapClaims{"user": "go-kit"}
	signedKey = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjoiZ28ta2l0In0.MMefQU5pwDeoWBSdyagqNlr1tDGddGUOMGiIWmMlFvk"
)

func TestJWTSigner(t *testing.T) {
	e := func(ctx context.Context, i interface{}) (interface{}, error) { return ctx, nil }

	signer := auth.NewJWTSigner(key, method, claims)(e)
	ctx := context.Background()
	ctx1, err := signer(ctx, struct{}{})
	if err != nil {
		t.Fatalf("Signer returned error: %s", err)
	}

	md, ok := metadata.FromContext(ctx1.(context.Context))
	if !ok {
		t.Fatal("Could not retrieve metadata from context")
	}

	token, ok := md["jwttoken"]
	if !ok {
		t.Fatal("Token did not exist in context")
	}

	if token[0] != signedKey {
		t.Fatalf("JWT tokens did not match: expecting %s got %s", signedKey, token[0])
	}
}

func TestJWTParser(t *testing.T) {
	e := func(ctx context.Context, i interface{}) (interface{}, error) { return ctx, nil }

	keyfunc := func(token *jwt.Token) (interface{}, error) { return []byte(key), nil }

	parser := auth.NewJWTParser(keyfunc, method)(e)
	ctx := context.WithValue(context.Background(), "jwtToken", signedKey)
	ctx1, err := parser(ctx, struct{}{})
	if err != nil {
		t.Fatalf("Parser returned error: %s", err)
	}

	cl, ok := ctx1.(context.Context).Value("jwtClaims").(jwt.MapClaims)
	if !ok {
		t.Fatal("Claims were not passed into context correctly")
	}

	if cl["user"] != claims["user"] {
		t.Fatalf("JWT Claims.user did not match: expecting %s got %s", claims["user"], cl["user"])
	}
}

func TestGRPCServerRequestFunc(t *testing.T) {
	md := metadata.Pairs("jwttoken", signedKey)
	ctx := context.Background()
	reqFunc := auth.NewGRPCServerRequestFunc()

	ctx = reqFunc(ctx, &md)
	token, ok := ctx.Value("jwtToken").(string)
	if !ok {
		t.Fatal("JWT Token not passed to context correctly")
	}

	if token != signedKey {
		t.Fatalf("JWT tokens did not match: expecting %s got %s", signedKey, token)
	}
}

func TestGRPCClientRequestFunc(t *testing.T) {
	md := metadata.Pairs("jwttoken", signedKey)
	ctx := metadata.NewContext(context.Background(), md)

	reqFunc := auth.NewGRPCClientRequestFunc()

	reqFunc(ctx, &md)
	token, ok := md["jwttoken"]
	if !ok {
		t.Fatal("JWT Token not passed to metadata correctly")
	}

	if token[0] != signedKey {
		t.Fatalf("JWT tokens did not match: expecting %s got %s", signedKey, token[0])
	}
}
