package log

import (
	"errors"
	"io"
	"sync"
	"sync/atomic"
)

// SwapLogger wraps another logger that may be safely replaced while other
// goroutines use the SwapLogger concurrently. The zero value for a SwapLogger
// will discard all log events without error.
//
// SwapLogger serves well as a package global logger that can be changed by
// importers.
type SwapLogger struct {
	logger atomic.Value
}

type loggerStruct struct {
	Logger
}

// Log implements the Logger interface by forwarding keyvals to the currently
// wrapped logger. It does not log anything if the wrapped logger is nil.
func (l *SwapLogger) Log(keyvals ...interface{}) error {
	s, ok := l.logger.Load().(loggerStruct)
	if !ok || s.Logger == nil {
		return nil
	}
	return s.Log(keyvals...)
}

// Swap replaces the currently wrapped logger with logger. Swap may be called
// concurrently with calls to Log from other goroutines.
func (l *SwapLogger) Swap(logger Logger) {
	l.logger.Store(loggerStruct{logger})
}

// SyncWriter synchronizes concurrent writes to an io.Writer.
type SyncWriter struct {
	mu sync.Mutex
	w  io.Writer
}

// NewSyncWriter returns a new SyncWriter. The returned writer is safe for
// concurrent use by multiple goroutines.
func NewSyncWriter(w io.Writer) *SyncWriter {
	return &SyncWriter{w: w}
}

// Write writes p to the underlying io.Writer. If another write is already in
// progress, the calling goroutine blocks until the SyncWriter is available.
func (w *SyncWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	n, err = w.w.Write(p)
	w.mu.Unlock()
	return n, err
}

// syncLogger provides concurrent safe logging for another Logger.
type syncLogger struct {
	mu     sync.Mutex
	logger Logger
}

// NewSyncLogger returns a logger that synchronizes concurrent use of the
// wrapped logger. When multiple goroutines use the SyncLogger concurrently
// only one goroutine will be allowed to log to the wrapped logger at a time.
// The other goroutines will block until the logger is available.
func NewSyncLogger(logger Logger) Logger {
	return &syncLogger{logger: logger}
}

// Log logs keyvals to the underlying Logger. If another log is already in
// progress, the calling goroutine blocks until the syncLogger is available.
func (l *syncLogger) Log(keyvals ...interface{}) error {
	l.mu.Lock()
	err := l.logger.Log(keyvals...)
	l.mu.Unlock()
	return err
}

// AsyncLogger provides buffered asynchronous and concurrent safe logging for
// another logger.
//
// If the wrapped logger's Log method ever returns an error, the AsyncLogger
// will stop processing log events and make the error available via the Err
// method. Any unprocessed log events in the buffer will be lost.
type AsyncLogger struct {
	logger   Logger
	keyvalsC chan []interface{}

	stopping chan struct{}
	stopped  chan struct{}

	mu  sync.Mutex
	err error
}

// NewAsyncLogger returns a new AsyncLogger that logs to logger and can buffer
// up to size log events before overflowing.
func NewAsyncLogger(logger Logger, size int) *AsyncLogger {
	l := &AsyncLogger{
		logger:   logger,
		keyvalsC: make(chan []interface{}, size),
		stopping: make(chan struct{}),
		stopped:  make(chan struct{}),
	}
	go l.run()
	return l
}

// run forwards log events from l.keyvalsC to l.logger.
func (l *AsyncLogger) run() {
	defer close(l.stopped)
	for keyvals := range l.keyvalsC {
		err := l.logger.Log(keyvals...)
		if err != nil {
			l.mu.Lock()
			l.stop(err)
			l.mu.Unlock()
			return
		}
	}
}

// caller must hold l.mu
func (l *AsyncLogger) stop(err error) {
	if err != nil && l.err == nil {
		l.err = err
	}
	select {
	case <-l.stopping:
		// already stopping, do nothing
	default:
		close(l.stopping)
		close(l.keyvalsC)
	}
}

// Log queues keyvals for logging by the wrapped Logger. Log may be called
// concurrently by multiple goroutines. If the the buffer is full, Log will
// return ErrAsyncLoggerOverflow and the keyvals are not queued. If the
// AsyncLogger is stopping, Log will return ErrAsyncLoggerStopping.
func (l *AsyncLogger) Log(keyvals ...interface{}) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	select {
	case <-l.stopping:
		return ErrAsyncLoggerStopping
	default:
	}

	select {
	case l.keyvalsC <- keyvals:
		return nil
	default:
		return ErrAsyncLoggerOverflow
	}
}

// Errors returned by AsyncLogger.
var (
	ErrAsyncLoggerStopping = errors.New("aysnc logger: logger stopped")
	ErrAsyncLoggerOverflow = errors.New("aysnc logger: log buffer overflow")
)

// Stop stops the AsyncLogger. After stop returns the logger will not accept
// new log events. Log events queued prior to calling Stop will be logged.
func (l *AsyncLogger) Stop() {
	l.mu.Lock()
	l.stop(nil)
	l.mu.Unlock()
}

// Stopping returns a channel that is closed after Stop is called.
func (l *AsyncLogger) Stopping() <-chan struct{} {
	return l.stopping
}

// Stopped returns a channel that is closed after Stop is called and all log
// events have been sent to the wrapped logger.
func (l *AsyncLogger) Stopped() <-chan struct{} {
	return l.stopped
}

// Err returns the first error returned by the wrapped logger.
func (l *AsyncLogger) Err() error {
	l.mu.Lock()
	err := l.err
	l.mu.Unlock()
	return err
}
