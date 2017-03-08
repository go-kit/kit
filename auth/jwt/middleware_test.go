package jwt

import (
	"context"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/endpoint"
)

var (
	kid            = "kid"
	key            = []byte("test_signing_key")
	method         = jwt.SigningMethodHS256
	invalidMethod  = jwt.SigningMethodRS256
	claims         = Claims{"user": "go-kit"}
	mapClaims      = jwt.MapClaims{"user": "go-kit"}
	standardClaims = jwt.StandardClaims{Audience: "go-kit"}
	// Signed tokens generated at https://jwt.io/
	signedKey         = "eyJhbGciOiJIUzI1NiIsImtpZCI6ImtpZCIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjoiZ28ta2l0In0.14M2VmYyApdSlV_LZ88ajjwuaLeIFplB8JpyNy0A19E"
	standardSignedKey = "eyJhbGciOiJIUzI1NiIsImtpZCI6ImtpZCIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJnby1raXQifQ.L5ypIJjCOOv3jJ8G5SelaHvR04UJuxmcBN5QW3m_aoY"
	invalidKey        = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.e30.vKVCKto-Wn6rgz3vBdaZaCBGfCBDTXOENSo_X2Gq7qA"
)

func signingValidator(t *testing.T, signer endpoint.Endpoint, expectedKey string) {
	ctx, err := signer(context.Background(), struct{}{})
	if err != nil {
		t.Fatalf("Signer returned error: %s", err)
	}

	token, ok := ctx.(context.Context).Value(JWTTokenContextKey).(string)
	if !ok {
		t.Fatal("Token did not exist in context")
	}

	if token != expectedKey {
		t.Fatalf("JWT tokens did not match: expecting %s got %s", expectedKey, token)
	}
}

func TestNewSigner(t *testing.T) {
	e := func(ctx context.Context, i interface{}) (interface{}, error) { return ctx, nil }

	signer := NewSigner(kid, key, method, claims)(e)
	signingValidator(t, signer, signedKey)

	signer = NewSignerWithClaims(kid, key, method, mapClaims)(e)
	signingValidator(t, signer, signedKey)

	signer = NewSignerWithClaims(kid, key, method, standardClaims)(e)
	signingValidator(t, signer, standardSignedKey)
}

func TestJWTParser(t *testing.T) {
	e := func(ctx context.Context, i interface{}) (interface{}, error) { return ctx, nil }

	keys := func(token *jwt.Token) (interface{}, error) {
		return key, nil
	}

	parser := NewParser(keys, method)(e)

	// No Token is passed into the parser
	_, err := parser(context.Background(), struct{}{})
	if err == nil {
		t.Error("Parser should have returned an error")
	}

	if err != ErrTokenContextMissing {
		t.Errorf("unexpected error returned, expected: %s got: %s", ErrTokenContextMissing, err)
	}

	// Invalid Token is passed into the parser
	ctx := context.WithValue(context.Background(), JWTTokenContextKey, invalidKey)
	_, err = parser(ctx, struct{}{})
	if err == nil {
		t.Error("Parser should have returned an error")
	}

	// Invalid Method is used in the parser
	badParser := NewParser(keys, invalidMethod)(e)
	ctx = context.WithValue(context.Background(), JWTTokenContextKey, signedKey)
	_, err = badParser(ctx, struct{}{})
	if err == nil {
		t.Error("Parser should have returned an error")
	}

	if err != ErrUnexpectedSigningMethod {
		t.Errorf("unexpected error returned, expected: %s got: %s", ErrUnexpectedSigningMethod, err)
	}

	// Invalid key is used in the parser
	invalidKeys := func(token *jwt.Token) (interface{}, error) {
		return []byte("bad"), nil
	}

	badParser = NewParser(invalidKeys, method)(e)
	ctx = context.WithValue(context.Background(), JWTTokenContextKey, signedKey)
	_, err = badParser(ctx, struct{}{})
	if err == nil {
		t.Error("Parser should have returned an error")
	}

	// Correct token is passed into the parser
	ctx = context.WithValue(context.Background(), JWTTokenContextKey, signedKey)
	ctx1, err := parser(ctx, struct{}{})
	if err != nil {
		t.Fatalf("Parser returned error: %s", err)
	}

	cl, ok := ctx1.(context.Context).Value(JWTClaimsContextKey).(Claims)
	if !ok {
		t.Fatal("Claims were not passed into context correctly")
	}

	if cl["user"] != claims["user"] {
		t.Fatalf("JWT Claims.user did not match: expecting %s got %s", claims["user"], cl["user"])
	}

	parser = NewParserWithClaims(keys, method, &jwt.StandardClaims{})(e)
	ctx = context.WithValue(context.Background(), JWTTokenContextKey, standardSignedKey)
	ctx1, err = parser(ctx, struct{}{})
	if err != nil {
		t.Fatalf("Parser returned error: %s", err)
	}
	stdCl, ok := ctx1.(context.Context).Value(JWTClaimsContextKey).(*jwt.StandardClaims)
	if !ok {
		t.Fatal("Claims were not passed into context correctly")
	}
	if !stdCl.VerifyAudience("go-kit", true) {
		t.Fatal("JWT jwt.StandardClaims.Audience did not match: expecting %s got %s", standardClaims.Audience, stdCl.Audience)
	}
}
