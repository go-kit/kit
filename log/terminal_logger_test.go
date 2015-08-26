package log_test

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"testing"

	"github.com/go-kit/kit/log"
	"gopkg.in/logfmt.v0"
)

func TestTerminalLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := newTerminalLogger(t, &buf)

	if err := logger.Log("hello", "world"); err != nil {
		t.Fatal(err)
	}
	if want, have := "hello=world\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	if err := logger.Log("a", 1, "err", errors.New("error")); err != nil {
		t.Fatal(err)
	}
	if want, have := "a=1 err=error\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	if err := logger.Log("std_map", map[int]int{1: 2}, "my_map", mymap{0: 0}); err != nil {
		t.Fatal(err)
	}
	if want, have := "std_map=\""+logfmt.ErrUnsupportedValueType.Error()+"\" my_map=special_behavior\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}
}

func newTerminalLogger(t testing.TB, w io.Writer) log.Logger {
	log.IsTTY = func(_ uintptr) bool {
		return true
	}
	return log.NewTerminalLogger(w, log.NewLogfmtLogger(w), nil)
}

func BenchmarkTerminalLoggerSimple(b *testing.B) {
	benchmarkRunner(b, newTerminalLogger(b, ioutil.Discard), baseMessage)
}

func BenchmarkTerminalLoggerContextual(b *testing.B) {
	benchmarkRunner(b, newTerminalLogger(b, ioutil.Discard), withMessage)
}

func TestTerminalLoggerConcurrency(t *testing.T) {
	testConcurrency(t, newTerminalLogger(t, ioutil.Discard))
}
