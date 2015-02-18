package server

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"testing"
)

func TestOkay(t *testing.T) {
	codec := &AddCodecJSON{}
	service := AddService
	server := httptest.NewServer(HTTPService(codec, service))
	defer server.Close()

	var a, b int = 1, 2
	buf, err := json.Marshal(AddRequest{A: a, B: b})
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Post(server.URL, "application/json", bytes.NewReader(buf))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if want, have := http.StatusOK, resp.StatusCode; want != have {
		buf, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("want HTTP %d, have %d (%s)", want, have, buf)
	}

	var addResp AddResponse
	if err := json.NewDecoder(resp.Body).Decode(&addResp); err != nil {
		t.Fatal(err)
	}

	if want, have := (a + b), addResp.V; want != have {
		t.Fatalf("want %d, have %d", want, have)
	}
}
