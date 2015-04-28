#!/usr/bin/env sh

# Thrift code generation for Go is broken in the current stable (0.9.2)
# release. See https://issues.apache.org/jira/browse/THRIFT-3021. We prefix
# `thrift` as `_thrift` so the `go` tool ignores the subdir.
#
# See also
#  https://thrift.apache.org/tutorial/go

thrift -r --gen go:thrift_import=github.com/apache/thrift/lib/go/thrift add.thrift
