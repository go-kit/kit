package auth_test

import (
	"testing"

	"google.golang.org/grpc/metadata"

	jwt "github.com/dgrijalva/jwt-go"

	"github.com/briankassouf/kit/auth"
	"golang.org/x/net/context"
)

var (
	key       = "test_signing_key"
	method    = jwt.SigningMethodHS256
	signedKey = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.O_-KMcAh6-2m9J6skFIDOj0hJVoTi6QqzbEkifV_I4Y"
)

func TestJWTSigner(t *testing.T) {
	e := func(ctx context.Context, i interface{}) (interface{}, error) { return ctx, nil }

	signer := auth.NewJWTSigner(key, method)(e)
	ctx := context.Background()
	ctx1, err := signer(ctx, struct{}{})
	if err != nil {
		t.Errorf("Signer returned error: %s", err)
	}

	md, ok := metadata.FromContext(ctx1.(context.Context))
	if !ok {
		t.Error("Could not retrieve metadata from context")
	}

	token, ok := md["jwttoken"]
	if !ok {
		t.Error("Token did not exist in context")
	}

	if token[0] != signedKey {
		t.Errorf("JWT tokens did not match: expecting %s got %s", signedKey, token[0])
	}
}

func TestJWTParser(t *testing.T) {
	e := func(ctx context.Context, i interface{}) (interface{}, error) { return ctx, nil }

	keyfunc := func(token *jwt.Token) (interface{}, error) { return []byte(key), nil }

	parser := auth.NewJWTParser(keyfunc, method)(e)
	ctx := context.WithValue(context.Background(), "jwtToken", signedKey)
	_, err := parser(ctx, struct{}{})
	if err != nil {
		t.Errorf("Parser returned error: %s", err)
	}
}
