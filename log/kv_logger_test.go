package log_test

import (
	"bytes"
	"testing"

	"github.com/peterbourgon/gokit/log"
)

func TestKVLogger(t *testing.T) {
	buf := bytes.Buffer{}
	logger := log.NewKVLogger(&buf)
	logger.Log("a")
	if want, have := "a\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	logger = logger.With(log.Field{Key: "level", Value: "DEBUG"})
	if err := logger.Log("b"); err != nil {
		t.Fatal(err)
	}
	if want, have := "level=DEBUG b\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	if err := logger.With(log.Field{Key: "count", Value: 123}).Log("c"); err != nil {
		t.Fatal(err)
	}
	if want, have := "level=DEBUG count=123 c\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	if err := logger.With(log.Field{Key: "m", Value: map[int]int{1: 2}}).Log("d"); err != nil {
		t.Fatal(err)
	}
	if want, have := "level=DEBUG m=map[1:2] d\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	if err := logger.With(log.Field{Key: "my", Value: myMap{0: 0}}).Log("e"); err != nil {
		t.Fatal(err)
	}
	if want, have := "level=DEBUG my=special-behavior e\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	if err := logger.With(log.Field{Key: "a", Value: 1}, log.Field{Key: "b", Value: 2}).Log("f"); err != nil {
		t.Fatal(err)
	}
	if want, have := "level=DEBUG a=1 b=2 f\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}
}

type myMap map[int]int

func (m myMap) String() string { return "special-behavior" }
