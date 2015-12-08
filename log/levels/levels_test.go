package levels_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/levels"
)

func TestDefaultLevels(t *testing.T) {
	buf := bytes.Buffer{}
	logger := levels.New(log.NewLogfmtLogger(&buf))

	logger.Debug("key", "val").Log("msg", "résumé") // of course you'd want to do this
	if want, have := "level=debug key=val msg=résumé\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	logger.Info("key", "val").Log("msg", "Åhus")
	if want, have := "level=info key=val msg=Åhus\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	logger.Error("key", "val").Log("msg", "© violation")
	if want, have := "level=error key=val msg=\"© violation\"\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	logger.Crit("key", "val").Log("msg", "	")
	if want, have := "level=crit key=val msg=\"\\t\"\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}
}

func TestModifiedLevels(t *testing.T) {
	buf := bytes.Buffer{}
	logger := levels.New(
		log.NewJSONLogger(&buf),
		levels.Key("l"),
		levels.DebugValue("dbg"),
		levels.InfoValue("nfo"),
		levels.WarnValue("wrn"),
		levels.ErrorValue("err"),
		levels.CritValue("crt"),
	)
	logger.With("easter_island", "176°").Debug().Log("msg", "moai")
	if want, have := `{"easter_island":"176°","l":"dbg","msg":"moai"}`+"\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}
}

func ExampleLevels() {
	logger := levels.New(log.NewLogfmtLogger(os.Stdout))
	logger.Debug().Log("msg", "hello")
	logger.With("context", "foo").Warn().Log("err", "error")

	// Output:
	// level=debug msg=hello
	// level=warn context=foo err=error
}
