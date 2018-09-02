package log_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/go-kit/kit/log"
)

func TestLogrusLogger(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	logger := log.NewLogrusLogger(buf)

	if err := logger.Log("hello", "world"); err != nil {
		t.Fatal(err)
	}
	if want, have := "hello=world\n", strings.Split(buf.String(), " ")[3]; want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	if err := logger.Log("a", 1, "err", errors.New("error")); err != nil {
		t.Fatal(err)
	}
	if want, have := "a=1 err=error", strings.TrimSpace(strings.SplitAfterN(buf.String(), " ", 4)[3]); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	if err := logger.Log("my_map", mymap{0: 0}); err != nil {
		t.Fatal(err)
	}
	if want, have := "my_map=special_behavior", strings.TrimSpace(strings.Split(buf.String(), " ")[3]); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}
}

func BenchmarkLogrusLoggerSimple(b *testing.B) {
	benchmarkRunner(b, log.NewLogrusLogger(ioutil.Discard), baseMessage)
}

func BenchmarkLogrusLoggerContextual(b *testing.B) {
	benchmarkRunner(b, log.NewLogrusLogger(ioutil.Discard), withMessage)
}

func TestLogrusLoggerConcurrency(t *testing.T) {
	t.Parallel()
	testConcurrency(t, log.NewLogrusLogger(ioutil.Discard), 10000)
}
