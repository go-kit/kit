package log_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"testing"

	"github.com/go-kit/kit/log"
)

func TestJSONLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := log.NewJSONLogger(buf)
	if err := logger.Log("err", errors.New("err"), "m", map[string]int{"0": 0}, "a", []int{1, 2, 3}); err != nil {
		t.Fatal(err)
	}
	if want, have := `{"a":[1,2,3],"err":"err","m":{"0":0}}`+"\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}
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
