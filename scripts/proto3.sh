#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

cd "$1"

# protoc-gen-go
if [ ! -x "$(command -v protoc-gen-go)" ]; then
    echo "Installing google.golang.org/protobuf/cmd/protoc-gen-go"
    go install google.golang.org/protobuf/cmd/protoc-gen-go
fi 

echo "protoc-gen-go"
protoc --go_out=. *.proto

# protoc-gen-go-grpc
if [ ! -x "$(command -v protoc-gen-go-grpc)" ]; then
    echo "Installing google.golang.org/grpc/cmd/protoc-gen-go-grpc"
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
fi 

echo "protoc-gen-go-grpc"
protoc --go-grpc_out=. *.proto

# protoc-gen-tstypes
if [ ! -x "$(command -v protoc-gen-tstypes)" ]; then
    echo "Installing github.com/tmc/grpcutil/protoc-gen-tstypes"
    go install github.com/tmc/grpcutil/protoc-gen-tstypes
fi
echo "protoc-gen-tstypes"
protoc --tstypes_out=. --tstypes_opt=declare_namespace=false,outpattern="{{.BaseName}}.d.ts" *.proto

# protoc-gen-tsipc
if [ ! -x "$(command -v protoc-gen-tsipc)" ]; then
    echo "Installing github.com/gabriel/grpcutil/protoc-gen-tsipc"
    go install github.com/gabriel/grpcutil/protoc-gen-tsipc
fi
echo "protoc-gen-tsipc"    
protoc --tsipc_out=. *.proto
