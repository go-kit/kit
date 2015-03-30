package log_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"testing"

	"github.com/peterbourgon/gokit/log"
)

func TestPrefixLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := log.NewPrefixLogger(buf)

	if err := logger.Log("hello"); err != nil {
		t.Fatal(err)
	}
	if want, have := "hello\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	if err := logger.Log("world", "k", "v"); err != nil {
		t.Fatal(err)
	}
	if want, have := "k=v world\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	if err := logger.With("z", 1, "a", 2).Log("🐰", "m", errors.New("n")); err != nil {
		t.Fatal(err)
	}
	if want, have := "z=1 a=2 m=n 🐰\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}
}

func BenchmarkPrefixLoggerSimple(b *testing.B) {
	benchmarkRunner(b, log.NewPrefixLogger(ioutil.Discard), simpleMessage)
}

func BenchmarkPrefixLoggerContextual(b *testing.B) {
	benchmarkRunner(b, log.NewPrefixLogger(ioutil.Discard), contextualMessage)
}
