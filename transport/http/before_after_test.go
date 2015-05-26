package http_test

import (
	"net/http/httptest"
	"testing"

	"golang.org/x/net/context"

	httptransport "github.com/go-kit/kit/transport/http"
)

func TestSetContentType(t *testing.T) {
	contentType := "application/whatever"
	rec := httptest.NewRecorder()
	httptransport.SetContentType(contentType)(context.Background(), rec)
	if want, have := contentType, rec.Header().Get("Content-Type"); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}
