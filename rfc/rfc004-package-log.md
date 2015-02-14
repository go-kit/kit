---
RFC: 004
Author: Peter Bougon <peter@bourgon.org> & Brian Knox <bknox@digitalocean.com>
Status: Draft
---

# package log

## Motivation

http://peter.bourgon.org/go-kit/#package-log

## Scope

* Log SHALL be drop in compatible with the standard library logger.
* Log SHALL use key / value pairs for structured data.
* Log MAY allow multiple stored logging contexts each with a different set of k/v pairs.
* Log SHALL be transport agnostic by accepting and io.Writer as output target.
* Log SHALL provide a default stdout io.Writer.
* Log MAY implement io.MultiWriters allowing broadcast of logs over multiple transports.
* Log MAY provide configurable back pressure handling strategies in the case of blocked io.Writers
* Log MAY provide some built in transports such as syslog and logstash.
* Log SHALL be format agnostic by providing an interface for log formatting.
* Log MAY provide some built in formatters such as RFC3164, JSON, etc.
* Log SHALL provide a set of defined severity levels.
* Log MUST NOT intrinsicly tie severity levels to program actions - e.g., a call to a specific log level should not call a panic.
* Log MAY allow tying program actions such as panic to a log level as a configurable option via hooks.

## Implementation

* The initial implementation should be a minimal feature set focused only on the scope of the RFC.
* Additional features and niceties may be added as there is demonstrable proof that the features solve real world problems.
