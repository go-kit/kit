package log_test

import (
	"bytes"
	"testing"

	"github.com/peterbourgon/gokit/log"
)

func TestWith(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := log.NewJSONLogger(buf)
	logger = log.With(logger, "a", 123)
	logger = log.With(logger, "b", "c") // With should stack
	if err := logger.Log("msg", "message"); err != nil {
		t.Fatal(err)
	}
	if want, have := `{"a":123,"b":"c","msg":"message"}`+"\n", buf.String(); want != have {
		t.Errorf("want\n\t%#v, have\n\t%#v", want, have)
	}
}

func TestWither(t *testing.T) {
	logger := &mylogger{}
	log.With(logger, "a", "b").Log("c", "d")
	if want, have := 1, logger.withs; want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

type mylogger struct{ withs int }

func (l *mylogger) Log(keyvals ...interface{}) error { return nil }

func (l *mylogger) With(keyvals ...interface{}) log.Logger { l.withs++; return l }
