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

	logger.Debug("msg", "résumé") // of course you'd want to do this
	if want, have := "level=debug msg=résumé\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	logger.Info("msg", "Åhus")
	if want, have := "level=info msg=Åhus\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	logger.Error("msg", "© violation")
	if want, have := "level=error msg=\"© violation\"\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	logger.Crit("msg", "	")
	if want, have := "level=crit msg=\"\\t\"\n", buf.String(); want != have {
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
	logger.With("easter_island", "176°").Debug("msg", "moai")
	if want, have := `{"easter_island":"176°","l":"dbg","msg":"moai"}`+"\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}
}

func ExampleLevels() {
	logger := levels.New(log.NewLogfmtLogger(os.Stdout))
	logger.Debug("msg", "hello")
	logger.With("context", "foo").Warn("err", "error")

	// Output:
	// level=debug msg=hello
	// level=warn context=foo err=error
}
