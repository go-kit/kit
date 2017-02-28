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

	logger.Debug().Log("msg", "résumé") // of course you'd want to do this
	if want, have := "level=debug msg=résumé\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	logger.Info().Log("msg", "Åhus")
	if want, have := "level=info msg=Åhus\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	logger.Error().Log("msg", "© violation")
	if want, have := "level=error msg=\"© violation\"\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	logger.Crit().Log("msg", "	")
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
	logger.With("easter_island", "176°").Debug().Log("msg", "moai")
	if want, have := `{"easter_island":"176°","l":"dbg","msg":"moai"}`+"\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}
}

func TestFilteredLevels(t *testing.T) {
	buf := bytes.Buffer{}
	logger := levels.New(
		log.NewLogfmtLogger(&buf),
		levels.FilterLogLevel(levels.Warn),
	)

	debugValuerCalled := false
	logger.Debug().Log("msg", "a debug log", "other", log.Valuer(func() interface{} { debugValuerCalled = true; return 42 }))
	if want, have := "", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}
	if debugValuerCalled {
		t.Error("Evaluated valuer in a filtered debug log unnecessarily")
	}

	buf.Reset()
	infoValuerCalled := false
	logger.Info().Log("msg", "an info log", "other", log.Valuer(func() interface{} { debugValuerCalled = true; return 42 }))
	if want, have := "", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}
	if infoValuerCalled {
		t.Error("Evaluated valuer in a filtered debug log unnecessarily")
	}

	buf.Reset()
	logger.Warn().Log("msg", "a warning log")
	if want, have := "level=warn msg=\"a warning log\"\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	logger.Error().Log("msg", "an error log")
	if want, have := "level=error msg=\"an error log\"\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	logger.Crit().Log("msg", "a critical log")
	if want, have := "level=crit msg=\"a critical log\"\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}
}

func TestErrorPromotion(t *testing.T) {
	buf := bytes.Buffer{}
	logger := levels.New(
		log.NewLogfmtLogger(&buf),
		levels.PromoteErrors(true),
		levels.FilterLogLevel(levels.Error),
	)
	// Should promote past filtering.
	logger.Debug().Log("error", "some error")
	if want, have := "level=error error=\"some error\"\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	// Should not promote if log level is already higher than the error level.
	buf.Reset()
	logger.Crit().Log("error", "some error")
	if want, have := "level=crit error=\"some error\"\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	logger = levels.New(
		log.NewLogfmtLogger(&buf),
		levels.PromoteErrors(true),
		levels.PromoteErrorToLevel(levels.Warn),
		levels.ErrorKey("err"),
	)
	// Should respect the configured ErrorKey
	logger.Debug().Log("error", "some error")
	if want, have := "level=debug error=\"some error\"\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	// Should promote to the configured level
	buf.Reset()
	logger.Debug().Log("err", "some error")
	if want, have := "level=warn err=\"some error\"\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	// Should treat nil errors as not an error
	buf.Reset()
	logger.Debug().Log("err", nil)
	if want, have := "level=debug err=null\n", buf.String(); want != have {
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
