#!/usr/bin/env sh

# Thrift code generation for Go is broken in the current stable (0.9.2)
# release. Leaving this stubbed out until the fix is released.
# https://issues.apache.org/jira/browse/THRIFT-3021

# https://thrift.apache.org/tutorial/go
for f in *.thrift ; do
	thrift -r --gen go:thrift_import=github.com/apache/thrift/lib/go/thrift $f
done

