package log_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"testing"

	"github.com/go-kit/kit/log"
)

func TestJSONLogger(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	logger := log.NewJSONLogger(buf)
	if err := logger.Log("err", errors.New("err"), "m", map[string]int{"0": 0}, "a", []int{1, 2, 3}); err != nil {
		t.Fatal(err)
	}
	if want, have := `{"a":[1,2,3],"err":"err","m":{"0":0}}`+"\n", buf.String(); want != have {
		t.Errorf("\nwant %#v\nhave %#v", want, have)
	}
}

func TestJSONLoggerMissingValue(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	logger := log.NewJSONLogger(buf)
	if err := logger.Log("k"); err != nil {
		t.Fatal(err)
	}
	if want, have := `{"k":"(MISSING)"}`+"\n", buf.String(); want != have {
		t.Errorf("\nwant %#v\nhave %#v", want, have)
	}
}

func TestJSONLoggerNilStringerKey(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := log.NewJSONLogger(buf)
	if err := logger.Log((*stringer)(nil), "v"); err != nil {
		t.Fatal(err)
	}
	if want, have := `{"NULL":"v"}`+"\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}
}

func TestJSONLoggerNilErrorValue(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := log.NewJSONLogger(buf)
	if err := logger.Log("err", (*stringError)(nil)); err != nil {
		t.Fatal(err)
	}
	if want, have := `{"err":null}`+"\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}
}

type stringer string

func (s stringer) String() string {
	return string(s)
}

type stringError string

func (s stringError) Error() string {
	return string(s)
}

func BenchmarkJSONLoggerSimple(b *testing.B) {
	benchmarkRunner(b, log.NewJSONLogger(ioutil.Discard), baseMessage)
}

func BenchmarkJSONLoggerContextual(b *testing.B) {
	benchmarkRunner(b, log.NewJSONLogger(ioutil.Discard), withMessage)
}

func TestJSONLoggerConcurrency(t *testing.T) {
	testConcurrency(t, log.NewJSONLogger(ioutil.Discard))
}
