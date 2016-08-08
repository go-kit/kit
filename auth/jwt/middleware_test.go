package jwt_test

import (
	"testing"

	stdjwt "github.com/dgrijalva/jwt-go"

	"github.com/go-kit/kit/auth/jwt"
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

	token, ok := ctx1.(context.Context).Value(jwt.JWTTokenContextKey).(string)
	if !ok {
		t.Fatal("Token did not exist in context")
	}

	if token != signedKey {
		t.Fatalf("JWT tokens did not match: expecting %s got %s", signedKey, token)
	}
}

func TestJWTParser(t *testing.T) {
	e := func(ctx context.Context, i interface{}) (interface{}, error) { return ctx, nil }

	keyfunc := func(token *stdjwt.Token) (interface{}, error) { return []byte(key), nil }

	parser := jwt.NewParser(keyfunc, method)(e)
	ctx := context.WithValue(context.Background(), jwt.JWTTokenContextKey, signedKey)
	ctx1, err := parser(ctx, struct{}{})
	if err != nil {
		t.Fatalf("Parser returned error: %s", err)
	}

	cl, ok := ctx1.(context.Context).Value(jwt.JWTClaimsContextKey).(stdjwt.MapClaims)
	if !ok {
		t.Fatal("Claims were not passed into context correctly")
	}

	if cl["user"] != claims["user"] {
		t.Fatalf("JWT Claims.user did not match: expecting %s got %s", claims["user"], cl["user"])
	}
}
