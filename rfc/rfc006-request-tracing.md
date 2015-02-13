---
RFC: 006
Author: Peter Bourgon <peter@bourgon.org>
Status: Draft
---

# Request tracing

## Motivation

[Dapper][]-style request tracing is a necessary introspection tool in any large
distributed system. Gokit services should support request tracing, including
exposition of traces that are compatible with [Zipkin][].

[Dapper]: http://research.google.com/pubs/pub36356.html
[Zipkin]: http://itszero.github.io/blog/2014/03/03/introduction-to-twitters-zipkin

## Scope

- Request tracing SHALL use Dapper terminology: Trace, Span, Tree, etc.

- Regardless if request tracing is enabled or disabled, if incoming requests
  contain trace IDs, package server MUST transparently forward them to
  downstream services.

- If request tracing is enabled, and incoming requests do not contain trace
  IDs, package server SHALL generate relevant IDs and forward them to
  downstream services.

## Implementation

To be defined.

