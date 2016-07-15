// Package log provides a structured logger.
//
// Applications may want to produce logs to be consumed later, either by
// humans or machines. Humans might be interested in debugging errors, or
// tracing specific requests. Machines might be interested in counting
// interesting events, or aggregating information for offline processing. In
// both cases, it's important that the log messages be structured and
// actionable. Package log is designed to encourage both of these best
// practices.
//
// Basic Usage
//
// The fundamental interface is Logger. Loggers create log events from
// key/value data.
//
// Concurrent Safety
//
// Applications with multiple goroutines want each log event written to the
// same logger to remain separate from other log events. Package log provides
// multiple solutions for concurrent safe logging.
package log
