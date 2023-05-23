package jwt

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"crypto/subtle"

	"github.com/go-kit/kit/endpoint"
	"github.com/golang-jwt/jwt/v5"
)

type customClaims struct {
	MyProperty string `json:"my_property"`
	jwt.RegisteredClaims
}

func (c customClaims) VerifyMyProperty(p string) bool {
	return subtle.ConstantTimeCompare([]byte(c.MyProperty), []byte(p)) != 0
}

var (
	kid              = "kid"
	key              = []byte("test_signing_key")
	myProperty       = "some value"
	method           = jwt.SigningMethodHS256
	invalidMethod    = jwt.SigningMethodRS256
	mapClaims        = jwt.MapClaims{"user": "go-kit"}
	registeredClaims = jwt.RegisteredClaims{Audience: []string{"go-kit"}}
	myCustomClaims   = customClaims{MyProperty: myProperty, RegisteredClaims: registeredClaims}
	// Signed tokens generated at https://jwt.io/
	signedKey           = "eyJhbGciOiJIUzI1NiIsImtpZCI6ImtpZCIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjoiZ28ta2l0In0.14M2VmYyApdSlV_LZ88ajjwuaLeIFplB8JpyNy0A19E"
	registeredSignedKey = "eyJhbGciOiJIUzI1NiIsImtpZCI6ImtpZCIsInR5cCI6IkpXVCJ9.eyJhdWQiOlsiZ28ta2l0Il19.vqB-qPpEqKyEYqNsDsM7ZrWYG7ZEhJLwBXMzR0H3ajo"
	customSignedKey     = "eyJhbGciOiJIUzI1NiIsImtpZCI6ImtpZCIsInR5cCI6IkpXVCJ9.eyJteV9wcm9wZXJ0eSI6InNvbWUgdmFsdWUiLCJhdWQiOlsiZ28ta2l0Il19.Yus4v91ScNgx6_zgLJVYofo2vpZziA_vds7WPWwwgbE"
	invalidKey          = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.e30.vKVCKto-Wn6rgz3vBdaZaCBGfCBDTXOENSo_X2Gq7qA"
	malformedKey        = "malformed.jwt.token"
)

func signingValidator(t *testing.T, signer endpoint.Endpoint, expectedKey string) {
	ctx, err := signer(context.Background(), struct{}{})
	if err != nil {
		t.Fatalf("Signer returned error: %s", err)
	}

	token, ok := ctx.(context.Context).Value(JWTContextKey).(string)
	if !ok {
		t.Fatal("Token did not exist in context")
	}

	if token != expectedKey {
		t.Fatalf("JWTs did not match: expecting %s got %s", expectedKey, token)
	}
}

func TestNewSigner(t *testing.T) {
	e := func(ctx context.Context, i interface{}) (interface{}, error) { return ctx, nil }

	signer := NewSigner(kid, key, method, mapClaims)(e)
	signingValidator(t, signer, signedKey)

	signer = NewSigner(kid, key, method, registeredClaims)(e)
	signingValidator(t, signer, registeredSignedKey)

	signer = NewSigner(kid, key, method, myCustomClaims)(e)
	signingValidator(t, signer, customSignedKey)
}

func TestJWTParser(t *testing.T) {
	e := func(ctx context.Context, i interface{}) (interface{}, error) { return ctx, nil }

	keys := func(token *jwt.Token) (interface{}, error) {
		return key, nil
	}

	parser := NewParser(keys, method, MapClaimsFactory)(e)

	// No Token is passed into the parser
	_, err := parser(context.Background(), struct{}{})
	if err == nil {
		t.Error("Parser should have returned an error")
	}

	if err != ErrTokenContextMissing {
		t.Errorf("unexpected error returned, expected: %s got: %s", ErrTokenContextMissing, err)
	}

	// Invalid Token is passed into the parser
	ctx := context.WithValue(context.Background(), JWTContextKey, invalidKey)
	_, err = parser(ctx, struct{}{})
	if err == nil {
		t.Error("Parser should have returned an error")
	}

	// Invalid Method is used in the parser
	badParser := NewParser(keys, invalidMethod, MapClaimsFactory)(e)
	ctx = context.WithValue(context.Background(), JWTContextKey, signedKey)
	_, err = badParser(ctx, struct{}{})
	if err == nil {
		t.Error("Parser should have returned an error")
	}

	if !errors.Is(err, ErrUnexpectedSigningMethod) {
		t.Errorf("unexpected error returned, expected: %s got: %s", ErrUnexpectedSigningMethod, err)
	}

	// Invalid key is used in the parser
	invalidKeys := func(token *jwt.Token) (interface{}, error) {
		return []byte("bad"), nil
	}

	badParser = NewParser(invalidKeys, method, MapClaimsFactory)(e)
	ctx = context.WithValue(context.Background(), JWTContextKey, signedKey)
	_, err = badParser(ctx, struct{}{})
	if err == nil {
		t.Error("Parser should have returned an error")
	}

	// Correct token is passed into the parser
	ctx = context.WithValue(context.Background(), JWTContextKey, signedKey)
	ctx1, err := parser(ctx, struct{}{})
	if err != nil {
		t.Fatalf("Parser returned error: %s", err)
	}

	cl, ok := ctx1.(context.Context).Value(JWTClaimsContextKey).(jwt.MapClaims)
	if !ok {
		t.Fatal("Claims were not passed into context correctly")
	}

	if cl["user"] != mapClaims["user"] {
		t.Fatalf("JWT Claims.user did not match: expecting %s got %s", mapClaims["user"], cl["user"])
	}

	// Test for malformed token error response
	parser = NewParser(keys, method, RegisteredClaimsFactory)(e)
	ctx = context.WithValue(context.Background(), JWTContextKey, malformedKey)
	ctx1, err = parser(ctx, struct{}{})
	if want, have := ErrTokenMalformed, err; want != have {
		t.Fatalf("Expected %+v, got %+v", want, have)
	}

	// Test for expired token error response
	parser = NewParser(keys, method, RegisteredClaimsFactory)(e)
	expired := jwt.NewWithClaims(method, jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Second * 100))})
	token, err := expired.SignedString(key)
	if err != nil {
		t.Fatalf("Unable to Sign Token: %+v", err)
	}
	ctx = context.WithValue(context.Background(), JWTContextKey, token)
	ctx1, err = parser(ctx, struct{}{})
	if want, have := ErrTokenExpired, err; want != have {
		t.Fatalf("Expected %+v, got %+v", want, have)
	}

	// Test for not activated token error response
	parser = NewParser(keys, method, RegisteredClaimsFactory)(e)
	notactive := jwt.NewWithClaims(method, jwt.RegisteredClaims{NotBefore: jwt.NewNumericDate(time.Now().Add(time.Second * 100))})
	token, err = notactive.SignedString(key)
	if err != nil {
		t.Fatalf("Unable to Sign Token: %+v", err)
	}
	ctx = context.WithValue(context.Background(), JWTContextKey, token)
	ctx1, err = parser(ctx, struct{}{})
	if want, have := ErrTokenNotActive, err; want != have {
		t.Fatalf("Expected %+v, got %+v", want, have)
	}

	// test valid registered claims token
	parser = NewParser(keys, method, RegisteredClaimsFactory)(e)
	ctx = context.WithValue(context.Background(), JWTContextKey, registeredSignedKey)
	ctx1, err = parser(ctx, struct{}{})
	if err != nil {
		t.Fatalf("Parser returned error: %s", err)
	}
	regCl, ok := ctx1.(context.Context).Value(JWTClaimsContextKey).(*jwt.RegisteredClaims)
	if !ok {
		t.Fatal("Claims were not passed into context correctly")
	}
	if len(regCl.Audience) != 1 || regCl.Audience[0] != registeredClaims.Audience[0] {
		t.Fatalf("JWT jwt.RegisteredClaims.Audience did not match: expecting %s got %s", registeredClaims.Audience, regCl.Audience)
	}

	// test valid customized claims token
	parser = NewParser(keys, method, func() jwt.Claims { return &customClaims{} })(e)
	ctx = context.WithValue(context.Background(), JWTContextKey, customSignedKey)
	ctx1, err = parser(ctx, struct{}{})
	if err != nil {
		t.Fatalf("Parser returned error: %s", err)
	}
	custCl, ok := ctx1.(context.Context).Value(JWTClaimsContextKey).(*customClaims)
	if !ok {
		t.Fatal("Claims were not passed into context correctly")
	}
	if len(custCl.Audience) != 1 || custCl.Audience[0] != registeredClaims.Audience[0] {
		t.Fatalf("JWT customClaims.Audience did not match: expecting %s got %s", registeredClaims.Audience, custCl.Audience)
	}
	if !custCl.VerifyMyProperty(myProperty) {
		t.Fatalf("JWT customClaims.MyProperty did not match: expecting %s got %s", myProperty, custCl.MyProperty)
	}
}

func TestIssue562(t *testing.T) {
	var (
		kf  = func(token *jwt.Token) (interface{}, error) { return []byte("secret"), nil }
		e   = NewParser(kf, jwt.SigningMethodHS256, MapClaimsFactory)(endpoint.Nop)
		key = JWTContextKey
		val = "eyJhbGciOiJIUzI1NiIsImtpZCI6ImtpZCIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjoiZ28ta2l0In0.14M2VmYyApdSlV_LZ88ajjwuaLeIFplB8JpyNy0A19E"
		ctx = context.WithValue(context.Background(), key, val)
	)
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			e(ctx, struct{}{}) // fatal error: concurrent map read and map write
		}()
	}
	wg.Wait()
}
