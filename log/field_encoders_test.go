package log_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/peterbourgon/gokit/log"
)

func TestFieldEqualsValue(t *testing.T) {
	for f, want := range map[log.Field]string{
		log.Field{Key: "", Value: ""}:                 "=",
		log.Field{Key: "", Value: nil}:                "=<nil>",
		log.Field{Key: "", Value: 0}:                  "=0",
		log.Field{Key: "k", Value: "v"}:               "k=v",
		log.Field{Key: "k", Value: errors.New("123")}: "k=123",
	} {
		var buf bytes.Buffer
		if err := log.KeyEqualsValue(&buf, f); err != nil {
			t.Errorf("%v: %v", f, err)
			continue
		}
		if have := buf.String(); want != have {
			t.Errorf("%v: want %q, have %q", f, want, have)
		}
	}
}

func TestValueOnly(t *testing.T) {
	for f, want := range map[log.Field]string{
		log.Field{Key: "", Value: ""}:                 "",
		log.Field{Key: "", Value: nil}:                "<nil>",
		log.Field{Key: "", Value: 0}:                  "0",
		log.Field{Key: "k", Value: "v"}:               "v",
		log.Field{Key: "k", Value: errors.New("123")}: "123",
	} {
		var buf bytes.Buffer
		if err := log.ValueOnly(&buf, f); err != nil {
			t.Errorf("%v: %v", f, err)
			continue
		}
		if have := buf.String(); want != have {
			t.Errorf("%v: want %q, have %q", f, want, have)
		}
	}
}

func TestJSON(t *testing.T) {
	for f, want := range map[log.Field]string{
		log.Field{Key: "", Value: ""}:   `{"":""}`,
		log.Field{Key: "", Value: nil}:  `{"":null}`,
		log.Field{Key: "", Value: 0}:    `{"":0}`,
		log.Field{Key: "k", Value: "v"}: `{"k":"v"}`,
		log.Field{Key: "k", Value: struct {
			V int `json:"v"`
		}{V: 123}}: `{"k":{"v":123}}`,
	} {
		var buf bytes.Buffer
		if err := log.JSON(&buf, f); err != nil {
			t.Errorf("%v: %v", f, err)
			continue
		}
		if have := strings.TrimSpace(buf.String()); want != have {
			t.Errorf("%v: want %q, have %q", f, want, have)
		}
	}
}

func TestEncodeMany(t *testing.T) {
	var buf bytes.Buffer
	if err := log.EncodeMany(&buf, log.JSON, []log.Field{
		log.Field{Key: "k", Value: "v"},
		log.Field{Key: "123", Value: 456},
	}); err != nil {
		t.Fatal(err)
	}
	want := `{"k":"v"} {"123":456}`
	have := strings.TrimSpace(buf.String())
	if want != have {
		t.Fatalf("want %q, have %q", want, have)
	}
}
