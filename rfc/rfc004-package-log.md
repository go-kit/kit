---
RFC: 004
Authors: Brian Knox <bknox@digitalocean.com>
Status: Accepted
---

# package log

## Motivation

http://peter.bourgon.org/go-kit/#package-log

## Scope

### Key / Value Pairs
*	Log SHALL use key / value pairs for structured data.
*	Log MAY preserve type safety when expressing logs in a structured format that supports types
*	Log SHALL provide a way to set a default k/v set per log context
*	Log MAY provide some pre-canned default keys (level, time, etc) for convenience
*	Log MAY allow multiple stored logging contexts each with a different set of k/v pairs.
*	Log MAY allow per log call adhoc k/v pairs (see Logrus as an example)

### Transport
*	Log SHALL be transport agnostic with pluggable transports.
*	Log MAY implement io.MultiWriters allowing broadcast of logs over multiple transports.
*	Log MAY use channels instead
*	Log MAY provide some built in transports such as syslog and logstash.
*	Log MAY use encoding.* above the transport level
*	Log MAY provide configurable back pressure handling strategies in the case of blocked Writers

### Formats
*	Log SHALL be format agnostic by providing an interface for log formatting.
*	Log MAY provide some built in formatters such as RFC3164, JSON, etc.

### Levels
*	Log MAY provide a set of defined severity levels that can be used (perhaps as a wrapper).
*	Log SHALL include severity as a k/v pair and allow setting it through the same mechanism as any other k/v pair
*	Log MAY provide wrapper types as a convenience for setting the severity level
*	Log MUST NOT intrinsically tie severity levels to program actions - e.g., a call to a specific log level should not call a panic.
*	Log MAY allow tying program actions such as panic to a log level.

## Implementation

*	The initial implementation should be a minimal feature set focused only on the scope of the RFC.
*	Additional features and niceties may be added as there is demonstrable proof that the features solve real world problems.
