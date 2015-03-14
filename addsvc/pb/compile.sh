#!/usr/bin/env bash

# https://github.com/grpc/grpc-common/tree/master/go
protoc add.proto --go_out=plugins=grpc:.
