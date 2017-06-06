package log_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"sort"
	"strings"
	"testing"

	"github.com/go-kit/kit/log"
)

func TestLTSVLogger(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	logger := log.NewLTSVLogger(buf)
	if err := logger.Log("err", errors.New("err"), "m", map[string]int{"0": 0}, "a", []int{1, 2, 3}); err != nil {
		t.Fatal(err)
	}

	if want, have := "\n", buf.String(); !strings.HasSuffix(have, want) {
		t.Errorf("must end with %q, got %#v", want, have)
	}

	values := strings.Split(strings.TrimSpace(buf.String()), "\t")
	sort.Strings(values)

	if want, have := "a:[1 2 3]", values[0]; want != have {
		t.Errorf("\nwant %#v\nhave %#v", want, have)
	}

	if want, have := "err:err", values[1]; want != have {
		t.Errorf("\nwant %#v\nhave %#v", want, have)
	}

	if want, have := "m:map[0:0]", values[2]; want != have {
		t.Errorf("\nwant %#v\nhave %#v", want, have)
	}
}

func TestLTSVLoggerNilStringerKey(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := log.NewLTSVLogger(buf)
	if err := logger.Log((*stringer)(nil), "v"); err != nil {
		t.Fatal(err)
	}
	if want, have := "NULL:v\n", buf.String(); want != have {
		t.Errorf("\nwant %#v\nhave %#v", want, have)
	}
}

func TestLTSVLoggerNilErrorValue(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := log.NewLTSVLogger(buf)
	if err := logger.Log("err", (*stringError)(nil)); err != nil {
		t.Fatal(err)
	}
	if want, have := "err:\n", buf.String(); want != have {
		t.Errorf("\nwant %#v\nhave %#v", want, have)
	}
}

func BenchmarkLTSVLoggerSimple(b *testing.B) {
	benchmarkRunner(b, log.NewLTSVLogger(ioutil.Discard), baseMessage)
}

func BenchmarkLTSVLoggerContextual(b *testing.B) {
	benchmarkRunner(b, log.NewLTSVLogger(ioutil.Discard), withMessage)
}

func TestLTSVLoggerConcurrency(t *testing.T) {
	testConcurrency(t, log.NewLTSVLogger(ioutil.Discard))
}
