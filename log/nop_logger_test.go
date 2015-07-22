package log_test

import (
	"testing"

	"github.com/go-kit/kit/log"
)

func TestNopLogger(t *testing.T) {
	logger := log.NewNopLogger()
	if err := logger.Log("abc", 123); err != nil {
		t.Error(err)
	}
	if err := log.NewContext(logger).With("def", "ghi").Log(); err != nil {
		t.Error(err)
	}
}
