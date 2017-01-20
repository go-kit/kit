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

	// Wrapping a context with the level logger allows users to use
	// log.DefaultCaller as per normal.
	var logger log.Logger
	logger = log.NewLogfmtLogger(&buf)
	logger = log.NewContext(logger).With("caller", log.DefaultCaller)

	logger = level.AllowingAll(logger)
	level.Info(logger).Log("foo", "bar")
	if want, have := `caller=level_test.go:109 level=info foo=bar`, strings.TrimSpace(buf.String()); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}

func TestMoreRestrictiveLevelLayering(t *testing.T) {
	var buf bytes.Buffer
	logger := log.NewJSONLogger(&buf)
	logger = level.AllowingAll(logger)
	logger = level.AllowingInfoAndAbove(logger)
	level.Debug(logger).Log("this is", "debug log")
	if want, have := "", strings.TrimSpace(buf.String()); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}

func TestLessRestrictiveLevelLayering(t *testing.T) {
	var buf bytes.Buffer
	logger := log.NewJSONLogger(&buf)
	logger = level.AllowingInfoAndAbove(logger)
	logger = level.AllowingAll(logger)
	level.Debug(logger).Log("this is", "debug log")
	if want, have := "", strings.TrimSpace(buf.String()); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}
