package log_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"testing"

	"github.com/go-kit/kit/log"
)

func TestPrefixLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := log.NewPrefixLogger(buf)

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
	if want, have := "std_map=map[1:2] my_map=special_behavior\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}
}

func BenchmarkPrefixLoggerSimple(b *testing.B) {
	benchmarkRunner(b, log.NewPrefixLogger(ioutil.Discard), baseMessage)
}

func BenchmarkPrefixLoggerContextual(b *testing.B) {
	benchmarkRunner(b, log.NewPrefixLogger(ioutil.Discard), withMessage)
}

func TestPrefixLoggerConcurrency(t *testing.T) {
	testConcurrency(t, log.NewPrefixLogger(ioutil.Discard))
}

type mymap map[int]int

func (m mymap) String() string { return "special_behavior" }
