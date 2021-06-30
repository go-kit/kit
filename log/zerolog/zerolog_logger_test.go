package zerolog_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	adapter "github.com/go-kit/kit/log/zerolog"
	"github.com/rs/zerolog"
)

func TestLeveledZerologLogger(t *testing.T) {
	levelKey := fmt.Sprint(level.Key())

	// basic test cases
	type testCase struct {
		zerologLevel zerolog.Level
		level        func(logger kitlog.Logger) kitlog.Logger
		kvs          []interface{}
		want         map[string]string
	}

	testCases := []testCase{
		{
			zerologLevel: zerolog.DebugLevel, level: level.Debug,
			kvs:  []interface{}{"key1", "value1"},
			want: map[string]string{levelKey: "debug", "key1": "value1"},
		},

		{
			zerologLevel: zerolog.InfoLevel, level: level.Info,
			kvs:  []interface{}{"key2", "value2"},
			want: map[string]string{levelKey: "info", "key2": "value2"},
		},

		{
			zerologLevel: zerolog.WarnLevel, level: level.Warn,
			kvs:  []interface{}{"key3", "value3"},
			want: map[string]string{levelKey: "warn", "key3": "value3"},
		},

		{
			zerologLevel: zerolog.ErrorLevel, level: level.Error,
			kvs:  []interface{}{"key4", "value4"},
			want: map[string]string{levelKey: "error", "key4": "value4"},
		},
	}

	// test
	for _, testCase := range testCases {
		t.Run(testCase.zerologLevel.String(), func(t *testing.T) {
			// make logger
			writer := &tbWriter{tb: t}
			logger := zerolog.New(writer)
			kitLogger := adapter.NewZerologLogger(&logger)

			testCase.level(kitLogger).Log(testCase.kvs...)
			check(t, writer, testCase.want)
		})
	}
}

func TestNoLevelZerologLogger(t *testing.T) {
	levelKey := fmt.Sprint(level.Key())

	// basic test cases
	type testCase struct {
		kvs  []interface{}
		want map[string]string
	}

	testCases := []testCase{
		{
			kvs:  []interface{}{"key1", "value1"},
			want: map[string]string{levelKey: "debug", "key1": "value1"},
		},

		{
			kvs:  []interface{}{"msg", "value2"},
			want: map[string]string{levelKey: "debug", "message": "value2"},
		},

		{
			kvs:  []interface{}{"key3"},
			want: map[string]string{levelKey: "debug", "key3": "(MISSING)"},
		},

		{
			kvs:  []interface{}{},
			want: map[string]string{levelKey: "debug"},
		},
	}

	// test
	for i, testCase := range testCases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			// make logger
			writer := &tbWriter{tb: t}
			logger := zerolog.New(writer)
			kitLogger := adapter.NewZerologLogger(&logger)

			kitLogger.Log(testCase.kvs...)
			check(t, writer, testCase.want)
		})
	}
}

// Check log kvs.
func check(t *testing.T, writer *tbWriter, want map[string]string) {
	logMap := make(map[string]string)

	err := json.Unmarshal([]byte(writer.sb.String()), &logMap)
	if err != nil {
		t.Errorf("unmarshal error: %v", err)
	} else {
		for k, v := range want {
			vv, ok := logMap[k]
			if !ok || v != vv {
				t.Errorf("error log, wanted %v, got %v", v, vv)
			}
		}
	}
}

type tbWriter struct {
	tb testing.TB
	sb strings.Builder
}

func (w *tbWriter) Write(b []byte) (n int, err error) {
	w.tb.Logf(string(b))
	w.sb.Write(b)

	return len(b), nil
}
