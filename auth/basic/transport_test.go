package basic

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"google.golang.org/grpc/metadata"
)

var (
	basicAuthToken = "dXNlcm5hbWU6cGFzc3dvcmQ="
)

func TestHTTPToContext(t *testing.T) {
	reqFunc := HTTPToContext()

	// When the header doesn't exist
	ctx := reqFunc(context.Background(), &http.Request{})

	if ctx.Value(BasicTokenContextKey) != nil {
		t.Error("Context shouldn't contain the encoded token")
	}

	// Authorization header value has invalid format
	header := http.Header{}
	header.Set("Authorization", "no expected auth header format value")
	ctx = reqFunc(context.Background(), &http.Request{Header: header})

	if ctx.Value(BasicTokenContextKey) != nil {
		t.Error("Context shouldn't contain the encoded token")
	}

	// Authorization header is correct
	header.Set("Authorization", generateAuthHeaderFromToken(basicAuthToken))
	ctx = reqFunc(context.Background(), &http.Request{Header: header})

	token := ctx.Value(BasicTokenContextKey).(string)
	if token != basicAuthToken {
		t.Errorf("Context doesn't contain the expected encoded token value; expected: %s, got: %s", basicAuthToken, token)
	}
}

func TestContextToHTTP(t *testing.T) {
	reqFunc := ContextToHTTP()

	// No Token is passed in the context
	ctx := context.Background()
	r := http.Request{}
	reqFunc(ctx, &r)

	token := r.Header.Get("Authorization")
	if token != "" {
		t.Error("authorization key should not exist in metadata")
	}

	// Correct Token is passed in the context
	ctx = context.WithValue(context.Background(), BasicTokenContextKey, basicAuthToken)
	r = http.Request{Header: http.Header{}}
	reqFunc(ctx, &r)

	token = r.Header.Get("Authorization")
	expected := generateAuthHeaderFromToken(basicAuthToken)

	if token != expected {
		t.Errorf("Authorization header does not contain the expected token; expected %s, got %s", expected, token)
	}
}

func TestGRPCToContext(t *testing.T) {
	md := metadata.MD{}
	reqFunc := GRPCToContext()

	// No Authorization header is passed
	ctx := reqFunc(context.Background(), md)
	token := ctx.Value(BasicTokenContextKey)
	if token != nil {
		t.Error("Context should not contain a token")
	}

	// Invalid Authorization header is passed
	md["authorization"] = []string{fmt.Sprintf("%s", basicAuthToken)}
	ctx = reqFunc(context.Background(), md)
	token = ctx.Value(BasicTokenContextKey)
	if token != nil {
		t.Error("Context should not contain a token")
	}

	// Authorization header is correct
	md["authorization"] = []string{fmt.Sprintf("Basic %s", basicAuthToken)}
	ctx = reqFunc(context.Background(), md)
	token, ok := ctx.Value(BasicTokenContextKey).(string)
	if !ok {
		t.Fatal("Token not passed to context correctly")
	}

	if token != basicAuthToken {
		t.Errorf("Tokens did not match: expecting %s got %s", basicAuthToken, token)
	}
}

func TestContextToGRPC(t *testing.T) {
	reqFunc := ContextToGRPC()

	// No Token is passed in the context
	ctx := context.Background()
	md := metadata.MD{}
	reqFunc(ctx, &md)

	// Correct Token is passed in the context
	ctx = context.WithValue(context.Background(), BasicTokenContextKey, basicAuthToken)
	md = metadata.MD{}
	reqFunc(ctx, &md)

	token, ok := md["authorization"]
	if !ok {
		t.Fatal("Token not passed to metadata correctly")
	}

	if token[0] != generateAuthHeaderFromToken(basicAuthToken) {
		t.Errorf("Tokens did not match: expecting %s got %s", basicAuthToken, token[0])
	}
}
