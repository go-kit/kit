package log_test

import (
	"bytes"
	"testing"

	"github.com/go-kit/kit/log"
)

func TestDefaultLevels(t *testing.T) {
	buf := bytes.Buffer{}
	levels := log.NewLevels(log.NewLogfmtLogger(&buf))

	levels.Debug.Log("msg", "👨") // of course you'd want to do this
	if want, have := "level=DEBUG msg=👨\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	levels.Info.Log("msg", "🚀")
	if want, have := "level=INFO msg=🚀\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	levels.Error.Log("msg", "🍵")
	if want, have := "level=ERROR msg=🍵\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}
}

func TestModifiedLevels(t *testing.T) {
	buf := bytes.Buffer{}
	levels := log.NewLevels(
		log.NewJSONLogger(&buf),
		log.LevelKey("l"),
		log.DebugLevelValue("⛄"),
		log.InfoLevelValue("🌜"),
		log.ErrorLevelValue("🌊"),
	)
	log.With(levels.Debug, "easter_island", "🗿").Log("msg", "💃💃💃")
	if want, have := `{"easter_island":"🗿","l":"⛄","msg":"💃💃💃"}`+"\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}
}
