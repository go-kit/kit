package enabled_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/levels"
	"github.com/go-kit/kit/log/levels/enabled"
)

func TestInfoLevel(t *testing.T) {
	buf := bytes.Buffer{}
	logger := enabled.New(levels.New(log.NewLogfmtLogger(&buf)), enabled.Info())

	if logger.DebugEnabled() {
		logger.Debug().Log("msg", "résumé") // of course you'd want to do this
	}
	if want, have := "", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	if logger.InfoEnabled() {
		logger.Info().Log("msg", "Åhus")
	}
	if want, have := "level=info msg=Åhus\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	if logger.ErrorEnabled() {
		logger.Error().Log("msg", "© violation")
	}
	if want, have := "level=error msg=\"© violation\"\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	if logger.CritEnabled() {
		logger.Crit().Log("msg", "	")
	}
	if want, have := "level=crit msg=\"\\t\"\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}
}

func ExampleEnabled() {
	logger := enabled.New(levels.New(log.NewLogfmtLogger(os.Stdout)), enabled.Warn())
	if logger.DebugEnabled() {
		logger.Debug().Log("msg", "hello")
	}
	if logger.WarnEnabled() {
		logger.With("context", "foo").Warn().Log("err", "error")
	}

	// Output:
	// level=warn context=foo err=error
}
