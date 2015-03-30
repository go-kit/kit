package log_test

import (
	"bytes"
	"testing"

	"github.com/peterbourgon/gokit/log"
)

func TestBasicLevels(t *testing.T) {
	buf := bytes.Buffer{}
	levels := log.NewLevels(log.NewPrefixLogger(&buf))

	levels.Debug.Log("👨") // of course you'd want to do this
	if want, have := "level=DEBUG 👨\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	levels.Info.Log("🚀")
	if want, have := "level=INFO 🚀\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	buf.Reset()
	levels.Error.Log("🍵")
	if want, have := "level=ERROR 🍵\n", buf.String(); want != have {
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

	levels.Debug.With("easter_island", "🗿").Log("💃💃💃")
	if want, have := `{"easter_island":"🗿","l":"⛄","msg":"💃💃💃"}`+"\n", buf.String(); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}
}
