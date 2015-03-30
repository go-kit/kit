package log_test

import (
	"bytes"
	"io/ioutil"
	"testing"
	"time"

	"github.com/peterbourgon/gokit/log"
)

func TestJSONLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := log.NewJSONLogger(buf)

	logger.Log("a")
	if want, have := `{"msg":"a"}`+"\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	if err := logger.With("level", "INFO").Log("b"); err != nil {
		t.Fatal(err)
	}
	if want, have := `{"level":"INFO","msg":"b"}`+"\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	logger = logger.With("request.size", 1024)
	logger = logger.With("response.code", 200)
	logger = logger.With("response.duration", 42*time.Millisecond)
	logger = logger.With("headers", map[string][]string{"X-Foo": []string{"A", "B"}})
	if err := logger.Log("OK"); err != nil {
		t.Fatal(err)
	}
	if want, have := `{"headers":{"X-Foo":["A","B"]},"msg":"OK","request.size":1024,"response.code":200,"response.duration":42000000}`+"\n", buf.String(); want != have {
		t.Errorf("want\n\t%#v, have\n\t%#v", want, have)
	}
}

func BenchmarkJSONLoggerSimple(b *testing.B) {
	benchmarkRunner(b, log.NewJSONLogger(ioutil.Discard), simpleMessage)
}

func BenchmarkJSONLoggerContextual(b *testing.B) {
	benchmarkRunner(b, log.NewJSONLogger(ioutil.Discard), contextualMessage)
}
