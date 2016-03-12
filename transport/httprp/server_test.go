package httprp_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"golang.org/x/net/context"

	httptransport "github.com/go-kit/kit/transport/httprp"
)

func TestServerHappyPathSingleServer(t *testing.T) {
	originServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("hey"))
		}))
	defer originServer.Close()
	originURL, _ := url.Parse(originServer.URL)

	handler := httptransport.NewServer(
		context.Background(),
		originURL,
	)
	proxyServer := httptest.NewServer(handler)
	defer proxyServer.Close()

	resp, _ := http.Get(proxyServer.URL)
	if want, have := http.StatusOK, resp.StatusCode; want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	responseBody, _ := ioutil.ReadAll(resp.Body)
	if want, have := "hey", string(responseBody); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestServerHappyPathSingleServerWithServerOptions(t *testing.T) {
	originServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if want, have := "go-kit-proxy", r.Header.Get("X-TEST-HEADER"); want != have {
				t.Errorf("want %d, have %d", want, have)
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("hey"))
		}))
	defer originServer.Close()
	originURL, _ := url.Parse(originServer.URL)

	handler := httptransport.NewServer(
		context.Background(),
		originURL,
		httptransport.ServerBefore(func(ctx context.Context, r *http.Request) context.Context {
			r.Header.Add("X-TEST-HEADER", "go-kit-proxy")
			return ctx
		}),
	)
	proxyServer := httptest.NewServer(handler)
	defer proxyServer.Close()

	resp, _ := http.Get(proxyServer.URL)
	if want, have := http.StatusOK, resp.StatusCode; want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	responseBody, _ := ioutil.ReadAll(resp.Body)
	if want, have := "hey", string(responseBody); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestServerOriginServerNotFoundResponse(t *testing.T) {
	originServer := httptest.NewServer(http.NotFoundHandler())
	defer originServer.Close()
	originURL, _ := url.Parse(originServer.URL)

	handler := httptransport.NewServer(
		context.Background(),
		originURL,
	)
	proxyServer := httptest.NewServer(handler)
	defer proxyServer.Close()

	resp, _ := http.Get(proxyServer.URL)
	if want, have := http.StatusNotFound, resp.StatusCode; want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestServerOriginServerUnreachable(t *testing.T) {
	// create a server, then promptly shut it down
	originServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
	originURL, _ := url.Parse(originServer.URL)
	originServer.Close()

	handler := httptransport.NewServer(
		context.Background(),
		originURL,
	)
	proxyServer := httptest.NewServer(handler)
	defer proxyServer.Close()

	resp, _ := http.Get(proxyServer.URL)
	if want, have := http.StatusInternalServerError, resp.StatusCode; want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}
