// Package log provides a structured logger.
//
// Services produce logs to be consumed later, either by humans or machines.
// Humans might be interested in debugging errors, or tracing specific requests.
// Machines might be interested in counting interesting events, or aggregating
// information for offline processing. In both cases, it's important that the
// log messages be structured and actionable. Package log is designed to
// encourage both of these best practices.
package log
