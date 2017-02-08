package level_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/alt_experimental_level"
)

func TestInstanceLevels(t *testing.T) {
	for _, testcase := range []struct {
		allowed string
		allow   func(log.Logger) log.Logger
		want    string
	}{
		{
			"all",
			level.AllowingAll,
			strings.Join([]string{
				`{"level":"debug","this is":"debug log"}`,
				`{"level":"info","this is":"info log"}`,
				`{"level":"warn","this is":"warn log"}`,
				`{"level":"error","this is":"error log"}`,
			}, "\n"),
		},
		{
			"debug+",
			level.AllowingDebugAndAbove,
			strings.Join([]string{
				`{"level":"debug","this is":"debug log"}`,
				`{"level":"info","this is":"info log"}`,
				`{"level":"warn","this is":"warn log"}`,
				`{"level":"error","this is":"error log"}`,
			}, "\n"),
		},
		{
			"info+",
			level.AllowingInfoAndAbove,
			strings.Join([]string{
				`{"level":"info","this is":"info log"}`,
				`{"level":"warn","this is":"warn log"}`,
				`{"level":"error","this is":"error log"}`,
			}, "\n"),
		},
		{
			"warn+",
			level.AllowingWarnAndAbove,
			strings.Join([]string{
				`{"level":"warn","this is":"warn log"}`,
				`{"level":"error","this is":"error log"}`,
			}, "\n"),
		},
		{
			"error",
			level.AllowingErrorOnly,
			strings.Join([]string{
				`{"level":"error","this is":"error log"}`,
			}, "\n"),
		},
		{
			"none",
			level.AllowingNone,
			``,
		},
	} {
		var buf bytes.Buffer
		logger := testcase.allow(log.NewJSONLogger(&buf))

		level.Debug(logger).Log("this is", "debug log")
		level.Info(logger).Log("this is", "info log")
		level.Warn(logger).Log("this is", "warn log")
		level.Error(logger).Log("this is", "error log")

		if want, have := testcase.want, strings.TrimSpace(buf.String()); want != have {
			t.Errorf("given Allowed=%s: want\n%s\nhave\n%s", testcase.allowed, want, have)
		}
	}
}

func TestLevelContext(t *testing.T) {
	var buf bytes.Buffer

	// Wrapping the level logger with a context allows users to use
	// log.DefaultCaller as per normal.
	var logger log.Logger
	logger = log.NewLogfmtLogger(&buf)
	logger = level.AllowingAll(logger)
	logger = level.Info(logger)
	logger = log.NewContext(logger).With("caller", log.DefaultCaller)

	logger.Log("foo", "bar")
	if want, have := `level=info caller=level_test.go:93 foo=bar`, strings.TrimSpace(buf.String()); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}

func TestContextLevel(t *testing.T) {
	var buf bytes.Buffer

	// Wrapping the level logger with a context allows users to use
	// log.DefaultCaller as per normal.
	var logger log.Logger
	logger = log.NewLogfmtLogger(&buf)
	logger = log.NewContext(logger).With("caller", log.DefaultCaller)

	logger = level.AllowingAll(logger)
	level.Info(logger).Log("foo", "bar")
	if want, have := `level=info caller=level_test.go:109 foo=bar`, strings.TrimSpace(buf.String()); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}

func TestLevelLayerRestrictions(t *testing.T) {
	factories := []struct {
		name string
		f    func(log.Logger) log.Logger
	}{
		{"all", level.AllowingAll},
		{"debug+", level.AllowingDebugAndAbove},
		{"info+", level.AllowingInfoAndAbove},
		{"warn+", level.AllowingWarnAndAbove},
		{"error", level.AllowingErrorOnly},
		{"none", level.AllowingNone},
	}
	emitters := []struct {
		name string
		f    func(log.Logger) log.Logger
	}{
		{"debug", level.Debug},
		{"info", level.Info},
		{"warn", level.Warn},
		{"error", level.Error},
	}
	tests := [][][4]bool{
		// all
		{
			{true, true, true, true},     // all
			{true, true, true, true},     // debug+
			{false, true, true, true},    // info+
			{false, false, true, true},   // warn+
			{false, false, false, true},  // error
			{false, false, false, false}, // none
		},
		// debug+
		{
			{true, true, true, true},     // all
			{true, true, true, true},     // debug+
			{false, true, true, true},    // info+
			{false, false, true, true},   // warn+
			{false, false, false, true},  // error
			{false, false, false, false}, // none
		},
		// info+
		{
			{false, true, true, true},    // all
			{false, true, true, true},    // debug+
			{false, true, true, true},    // info+
			{false, false, true, true},   // warn+
			{false, false, false, true},  // error
			{false, false, false, false}, // none
		},
		// warn+
		{
			{false, false, true, true},   // all
			{false, false, true, true},   // debug+
			{false, false, true, true},   // info+
			{false, false, true, true},   // warn+
			{false, false, false, true},  // error
			{false, false, false, false}, // none
		},
		// error
		{
			{false, false, false, true},  // all
			{false, false, false, true},  // debug+
			{false, false, false, true},  // info+
			{false, false, false, true},  // warn+
			{false, false, false, true},  // error
			{false, false, false, false}, // none
		},
		// none
		{
			{false, false, false, false}, // all
			{false, false, false, false}, // debug+
			{false, false, false, false}, // info+
			{false, false, false, false}, // warn+
			{false, false, false, false}, // error
			{false, false, false, false}, // none
		},
	}
	var buf bytes.Buffer
	logger := log.NewLogfmtLogger(&buf)
	for i, test := range tests {
		t.Run(factories[i].name, func(t *testing.T) {
			initialLogger := factories[i].f(logger)
			if initialLogger == nil {
				t.Fatal("initial factory returned nil")
			}
			// Wrap with an intervening layer to confirm that
			// subsequent level restricting factories can see through
			// to the inner restriction.
			initialLogger = log.NewContext(initialLogger)
			for j, layer := range test {
				t.Run(factories[j].name, func(t *testing.T) {
					layeredLogger := factories[j].f(initialLogger)
					if layeredLogger == nil {
						t.Fatal("layering factory returned nil")
					}
					for k, expected := range layer {
						t.Run(emitters[k].name, func(t *testing.T) {
							defer buf.Reset()
							leveled := emitters[k].f(layeredLogger)
							if leveled == nil {
								t.Fatalf("leveled emitter function returned nil")
							}
							leveled.Log("m", "x")
							if buf.Len() > 0 {
								if !expected {
									t.Fatalf("want no output, have %q", buf.Bytes())
								}
							} else if expected {
								t.Fatal("want output, have none")
							}
						})
					}
				})
			}
		})
	}
}
