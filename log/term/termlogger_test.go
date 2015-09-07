package term_test

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/term"
	"gopkg.in/logfmt.v0"
)

func TestTerminalLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := newTerminalLogger(t, &buf)

	if err := logger.Log("hello", "world", "level", "info"); err != nil {
		t.Fatal(err)
	}
	if want, have := "[INFO] []  hello=world\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	if err := logger.Log("a", 1, "level", "error", "err", errors.New("error")); err != nil {
		t.Fatal(err)
	}
	if want, have := "\x1b[31;1m[EROR] []  a=1 err=error\n\x1b[39;49m", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	if err := logger.Log("level", "crit", "std_map", map[int]int{1: 2}); err != nil {
		t.Fatal(err)
	}
	if want, have := "\x1b[41;1m[CRIT] []  std_map=\""+logfmt.ErrUnsupportedValueType.Error()+"\"\n\x1b[39;49m", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}
}

func newTerminalLogger(t testing.TB, w io.Writer) log.Logger {
	return term.NewTerminalLogger(w, nil)
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
