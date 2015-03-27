package cors_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/peterbourgon/gokit/transport/http/cors"
)

func TestProperGet(t *testing.T) {
	resp := httptest.NewRecorder()

	cors.Middleware()(code(http.StatusNotFound)).ServeHTTP(resp, &http.Request{
		Method: "GET",
		Header: map[string][]string{"Origin": {"localhost"}},
	})

	if want, have := http.StatusNotFound, resp.Code; want != have {
		t.Fatalf("want %d, have %d", want, have)
	}

	if want, have := "*", resp.HeaderMap.Get("Access-Control-Allow-Origin"); want != have {
		t.Fatalf("want %q, have %q", want, have)
	}

	for _, header := range []string{"Origin", "Accept", "Accept-Encoding", "Authorization", "Content-Type"} {
		if headers := resp.HeaderMap.Get("Access-Control-Allow-Headers"); !strings.Contains(headers, header) {
			t.Fatalf("looking for %q, got %v", header, headers)
		}
	}
}

func TestGetWithPost(t *testing.T) {
	resp := httptest.NewRecorder()

	cors.Middleware()(code(http.StatusTeapot)).ServeHTTP(resp, &http.Request{
		Method: "POST",
		Header: map[string][]string{"Origin": {"localhost"}},
	})

	if want, have := http.StatusMethodNotAllowed, resp.Code; want != have {
		t.Fatalf("want %d, have %d", want, have)
	}
}

func TestOptionsGet(t *testing.T) {
	resp := httptest.NewRecorder()

	cors.Middleware()(code(http.StatusNotFound)).ServeHTTP(resp, &http.Request{
		Method: "OPTIONS",
		Header: map[string][]string{
			"Access-Control-Request-Method": {"GET"},
			"Origin":                        {"localhost"},
		},
	})

	if want, have := http.StatusOK, resp.Code; want != have {
		t.Fatalf("want %d, have %d", want, have)
	}

	if want, have := "*", resp.HeaderMap.Get("Access-Control-Allow-Origin"); want != have {
		t.Fatalf("want %q, have %q", want, have)
	}

	for _, header := range []string{"Origin", "Accept", "Accept-Encoding", "Authorization", "Content-Type"} {
		if headers := resp.HeaderMap.Get("Access-Control-Allow-Headers"); !strings.Contains(headers, header) {
			t.Fatalf("looking for %q, got %v", header, headers)
		}
	}
}

type code int

func (c code) ServeHTTP(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(int(c)) }
