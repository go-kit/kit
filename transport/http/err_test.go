package http_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"golang.org/x/net/context"

	httptransport "github.com/go-kit/kit/transport/http"
)

func TestClientEndpointEncodeError(t *testing.T) {
	var (
		sampleErr = errors.New("Oh no, an error")
		enc       = func(r *http.Request, request interface{}) error { return sampleErr }
		dec       = func(r *http.Response) (response interface{}, err error) { return nil, nil }
	)

	u := &url.URL{
		Scheme: "https",
		Host:   "localhost",
		Path:   "/does/not/matter",
	}

	c := httptransport.NewClient(
		"GET",
		u,
		enc,
		dec,
	)

	_, err := c.Endpoint()(context.Background(), nil)
	if err == nil {
		t.Fatal("err == nil")
	}

	e, ok := err.(httptransport.TransportError)
	if !ok {
		t.Fatal("err is not of type github.com/go-kit/kit/transport/http.Err")
	}

	if want, have := sampleErr, e.Err; want != have {
		t.Fatalf("want %v, have %v", want, have)
	}
}

func ExampleErrOutput() {
	sampleErr := errors.New("Oh no, an error")
	err := httptransport.TransportError{"Do", sampleErr}
	fmt.Println(err)
	// Output:
	// Do: Oh no, an error
}
