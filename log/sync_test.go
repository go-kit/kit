package log_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/go-kit/kit/log"
)

func TestSwapLogger(t *testing.T) {
	t.Parallel()
	var logger log.SwapLogger

	// Zero value does not panic or error.
	err := logger.Log("k", "v")
	if got, want := err, error(nil); got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	buf := &bytes.Buffer{}
	json := log.NewJSONLogger(buf)
	logger.Swap(json)

	if err := logger.Log("k", "v"); err != nil {
		t.Error(err)
	}
	if got, want := buf.String(), `{"k":"v"}`+"\n"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	buf.Reset()
	prefix := log.NewLogfmtLogger(buf)
	logger.Swap(prefix)

	if err := logger.Log("k", "v"); err != nil {
		t.Error(err)
	}
	if got, want := buf.String(), "k=v\n"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	buf.Reset()
	logger.Swap(nil)

	if err := logger.Log("k", "v"); err != nil {
		t.Error(err)
	}
	if got, want := buf.String(), ""; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestSwapLoggerConcurrency(t *testing.T) {
	testConcurrency(t, &log.SwapLogger{})
}

func TestSyncLoggerConcurrency(t *testing.T) {
	var w io.Writer
	w = &bytes.Buffer{}
	logger := log.NewLogfmtLogger(w)
	logger = log.NewSyncLogger(logger)
	testConcurrency(t, logger)
}

func TestSyncWriterConcurrency(t *testing.T) {
	var w io.Writer
	w = &bytes.Buffer{}
	w = log.NewSyncWriter(w)
	testConcurrency(t, log.NewLogfmtLogger(w))
}

func TestAsyncLoggerConcurrency(t *testing.T) {
	var w io.Writer
	w = &bytes.Buffer{}
	logger := log.NewLogfmtLogger(w)
	al := log.NewAsyncLogger(logger, 10)
	testConcurrency(t, al)
	al.Stop()
	<-al.Stopped()
}

func TestAsyncLoggerLogs(t *testing.T) {
	t.Parallel()
	output := [][]interface{}{}
	logger := log.LoggerFunc(func(keyvals ...interface{}) error {
		output = append(output, keyvals)
		return nil
	})

	const logcnt = 10
	al := log.NewAsyncLogger(logger, logcnt)

	for i := 0; i < logcnt; i++ {
		al.Log("key", i)
	}

	al.Stop()
	<-al.Stopping()

	if got, want := al.Log("key", "late"), log.ErrAsyncLoggerStopping; got != want {
		t.Errorf(`logger err: got "%v", want "%v"`, got, want)
	}

	<-al.Stopped()

	if got, want := al.Err(), error(nil); got != want {
		t.Errorf(`logger err: got "%v", want "%v"`, got, want)
	}

	if got, want := len(output), logcnt; got != want {
		t.Errorf("logged events: got %v, want %v", got, want)
	}

	for i, e := range output {
		if got, want := e[1], i; got != want {
			t.Errorf("log event mismatch, got %v, want %v", got, want)
		}
	}
}

func TestAsyncLoggerLogError(t *testing.T) {
	t.Parallel()
	const logcnt = 10
	const logBeforeError = logcnt / 2
	logErr := errors.New("log error")

	output := [][]interface{}{}
	logger := log.LoggerFunc(func(keyvals ...interface{}) error {
		output = append(output, keyvals)
		if len(output) == logBeforeError {
			return logErr
		}
		return nil
	})

	al := log.NewAsyncLogger(logger, logcnt)

	for i := 0; i < logcnt; i++ {
		al.Log("key", i)
	}

	<-al.Stopping()

	if got, want := al.Log("key", "late"), log.ErrAsyncLoggerStopping; got != want {
		t.Errorf(`log while stopping err: got "%v", want "%v"`, got, want)
	}

	<-al.Stopped()

	if got, want := al.Err(), logErr; got != want {
		t.Errorf(`logger err: got "%v", want "%v"`, got, want)
	}

	if got, want := len(output), logBeforeError; got != want {
		t.Errorf("logged events: got %v, want %v", got, want)
	}

	for i, e := range output {
		if got, want := e[1], i; got != want {
			t.Errorf("log event mismatch, got %v, want %v", got, want)
		}
	}
}

func TestAsyncLoggerOverflow(t *testing.T) {
	t.Parallel()
	var (
		output     = make(chan []interface{}, 10)
		loggerdone = make(chan struct{})
	)

	logger := log.LoggerFunc(func(keyvals ...interface{}) error {
		output <- keyvals
		<-loggerdone // block here to stall the AsyncLogger.run loop
		return nil
	})

	al := log.NewAsyncLogger(logger, 1)

	if got, want := al.Log("k", 1), error(nil); got != want {
		t.Errorf(`first log err: got "%v", want "%v"`, got, want)
	}

	<-output
	// Now we know the AsyncLogger.run loop has consumed the first log event
	// and will be stalled until loggerdone is closed.

	// This log event fills the buffer without error.
	if got, want := al.Log("k", 2), error(nil); got != want {
		t.Errorf(`second log err: got "%v", want "%v"`, got, want)
	}

	// Now we test for buffer overflow.
	if got, want := al.Log("k", 3), log.ErrAsyncLoggerOverflow; got != want {
		t.Errorf(`third log err: got "%v", want "%v"`, got, want)
	}

	al.Stop()
	<-al.Stopping()

	if got, want := al.Log("key", "late"), log.ErrAsyncLoggerStopping; got != want {
		t.Errorf(`log while stopping err: got "%v", want "%v"`, got, want)
	}

	// Release the AsyncLogger.run loop and wait for it to stop.
	close(loggerdone)
	<-al.Stopped()

	if got, want := al.Err(), error(nil); got != want {
		t.Errorf(`logger err: got "%v", want "%v"`, got, want)
	}
}
