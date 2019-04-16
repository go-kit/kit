package log_test

import (
	"errors"
	"testing"

	"github.com/go-kit/kit/log"
)

func TestErrorHandler(t *testing.T) {
	var output []interface{}

	logger := log.Logger(log.LoggerFunc(func(keyvals ...interface{}) error {
		output = keyvals
		return nil
	}))

	errorHandler := log.NewErrorHandler(logger)

	err := errors.New("error")

	errorHandler.Handle(err)

	if output[1] != err {
		t.Errorf("expected an error log event: have %v, want %v", output[1], err)
	}
}
