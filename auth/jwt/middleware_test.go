package jwt_test

import (
	"testing"

	"google.golang.org/grpc/metadata"

	stdjwt "github.com/dgrijalva/jwt-go"

	"github.com/briankassouf/kit/auth/jwt"
	"golang.org/x/net/context"
)

var (
	key       = "test_signing_key"
	method    = stdjwt.SigningMethodHS256
	claims    = stdjwt.MapClaims{"user": "go-kit"}
	signedKey = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjoiZ28ta2l0In0.MMefQU5pwDeoWBSdyagqNlr1tDGddGUOMGiIWmMlFvk"
)

func TestSigner(t *testing.T) {
	e := func(ctx context.Context, i interface{}) (interface{}, error) { return ctx, nil }

	signer := jwt.NewSigner(key, method, claims)(e)
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

	keyfunc := func(token *stdjwt.Token) (interface{}, error) { return []byte(key), nil }

	parser := jwt.NewParser(keyfunc, method)(e)
	ctx := context.WithValue(context.Background(), "jwtToken", signedKey)
	ctx1, err := parser(ctx, struct{}{})
	if err != nil {
		t.Fatalf("Parser returned error: %s", err)
	}

	cl, ok := ctx1.(context.Context).Value("jwtClaims").(stdjwt.MapClaims)
	if !ok {
		t.Fatal("Claims were not passed into context correctly")
	}

	if cl["user"] != claims["user"] {
		t.Fatalf("JWT Claims.user did not match: expecting %s got %s", claims["user"], cl["user"])
	}
}
