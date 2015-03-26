package log_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/peterbourgon/gokit/log"
)

func TestBasicLogger(t *testing.T) {
	var (
		buf         bytes.Buffer
		baseLogger  = log.NewBasicLogger(&buf, log.ValueOnly)
		debugLogger = baseLogger.With(log.Field{Key: "level", Value: "DEBUG"})
		infoLogger  = baseLogger.With(log.Field{Key: "level", Value: "INFO"})
		errorLogger = baseLogger.With(log.Field{Key: "level", Value: "ERROR"})
	)

	debugLogger.Logf("debug %d", 1)
	infoLogger.Logf("info %d", 2)
	errorLogger.Logf("error %d", 3)

	if want, have := strings.Join([]string{
		`DEBUG debug 1`,
		`INFO info 2`,
		`ERROR error 3`,
	}, "\n"), strings.TrimSpace(buf.String()); want != have {
		t.Errorf("want \n%s, have \n%s", want, have)
	}
}
