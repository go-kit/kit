package http_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/net/context"

	transporthttp "github.com/go-kit/kit/transport/http"

	"testing"
)

func TestClientEndpointEncodeError(t *testing.T) {
	var (
		sampleErr = errors.New("Oh no, an error")
		enc       = func(r *http.Request, request interface{}) error {
			return sampleErr
		}
		dec = func(r *http.Response) (response interface{}, err error) {
			return nil, nil
		}
	)

	u := &url.URL{
		Scheme: "https",
		Host:   "localhost",
		Path:   "/does/not/matter",
	}

	c := transporthttp.NewClient(
		"GET",
		u,
		enc,
		dec,
	)

	_, err := c.Endpoint()(context.Background(), nil)
	if err == nil {
		t.Log("err == nil")
		t.Fail()
	}

	e, ok := err.(transporthttp.Err)
	if !ok {
		t.Log("err is not of type github.com/go-kit/kit/transport/http.Err")
		t.Fail()
	}

	if e.Err != sampleErr {
		t.Logf("e.Err != sampleErr, %s vs %s", e.Err, sampleErr)
		t.Fail()
	}
}

func ExampleErrOutput() {
	sampleErr := errors.New("Oh no, an error")
	err := transporthttp.Err{"Do", sampleErr}
	fmt.Println(err)
	// Output:
	// Do: Oh no, an error
}
