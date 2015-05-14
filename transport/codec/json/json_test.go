package json_test

import (
	"bytes"
	"testing"

	"golang.org/x/net/context"

	jsoncodec "github.com/go-kit/kit/transport/codec/json"
)

type request struct {
	A int    `json:"a"`
	B string `json:"b"`
}

type response struct {
	Values []string `json:"values"`
}

func TestDecode(t *testing.T) {
	buf := bytes.NewBufferString(`{"a":1,"b":"2"}`)
	var req request
	if _, err := jsoncodec.New().Decode(context.Background(), buf, &req); err != nil {
		t.Fatal(err)
	}
	if want, have := (request{A: 1, B: "2"}), req; want != have {
		t.Errorf("want %v, have %v", want, have)
	}
}

func TestEncode(t *testing.T) {
	buf := &bytes.Buffer{}
	if err := jsoncodec.New().Encode(buf, response{Values: []string{"a", "b", "c"}}); err != nil {
		t.Fatal(err)
	}
	if want, have := `{"values":["a","b","c"]}`+"\n", buf.String(); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}
