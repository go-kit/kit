package log_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/peterbourgon/gokit/log"
)

func TestJSONLoggerPrefixedFields(t *testing.T) {
	buf := bytes.Buffer{}
	logger := log.NewJSONLogger(&buf, log.PrefixedFields)
	if err := logger.Logf(`{"alpha" : "beta"}`); err != nil {
		t.Fatal(err)
	}
	if want, have := `{"alpha":"beta"}`+"\n", buf.String(); want != have {
		t.Fatalf("want\n\t%s, have\n\t%s", want, have)
	}

	buf.Reset()

	logger = logger.With(log.Field{Key: "foo", Value: "bar"})
	if err := logger.Logf(`{ "delta": "gamma" }`); err != nil {
		t.Fatal(err)
	}
	if want, have := `{"foo":"bar"} {"delta":"gamma"}`, strings.TrimSpace(buf.String()); want != have {
		t.Errorf("want\n\t%s, have\n\t%s", want, have)
	}
}

func TestJSONLoggerMixedFields(t *testing.T) {
	buf := bytes.Buffer{}
	logger := log.NewJSONLogger(&buf, log.MixedFields)
	logger = logger.With(log.Field{Key: "m", Value: "n"})
	if err := logger.Logf(`{"a":"b"}`); err != nil {
		t.Fatal(err)
	}
	if want, have := `{"a":"b","m":"n"}`, strings.TrimSpace(buf.String()); want != have {
		t.Fatalf("want\n\t%s, have\n\t%s", want, have)
	}
}
