#!/usr/bin/env sh

# Update protoc via
#  go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
#
# See also
#  https://github.com/grpc/grpc-common/tree/master/go

protoc add.proto --go_out=plugins=grpc:.
