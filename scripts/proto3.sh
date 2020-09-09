#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

cd "$1"

# protoc-gen-go
if [ ! -x "$(command -v protoc-gen-go)" ]; then
    echo "Installing google.golang.org/protobuf/cmd/protoc-gen-go"
    go install google.golang.org/protobuf/cmd/protoc-gen-go
fi 

# protoc-gen-go-grpc
if [ ! -x "$(command -v protoc-gen-go-grpc)" ]; then
    echo "Installing google.golang.org/grpc/cmd/protoc-gen-go-grpc"
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
fi

# protoc-gen-go-patch
if [ ! -x "$(command -v protoc-gen-go-patch)" ]; then
    echo "Installing github.com/alta/protopatch/cmd/protoc-gen-go-patch"
    go install github.com/alta/protopatch/cmd/protoc-gen-go-patch
fi

protoc_include="-I . \
  -I `go list -m -f {{.Dir}} github.com/alta/protopatch` \
  -I `go list -m -f {{.Dir}} google.golang.org/protobuf`"

echo "protoc-gen-go"
protoc \
  $protoc_include \
  --go-patch_out=plugin=go,paths=source_relative:. \
  --go-patch_out=plugin=go-grpc,paths=source_relative:. \
	*.proto

if [ ! -x "$(command -v protoc-gen-tstypes)" ]; then
    echo "Installing github.com/tmc/grpcutil/protoc-gen-tstypes"
    go install github.com/tmc/grpcutil/protoc-gen-tstypes
fi
echo "protoc-gen-tstypes"
protoc \
  $protoc_include \
  --tstypes_out=. \
  --tstypes_opt=declare_namespace=false,outpattern="{{.BaseName}}.d.ts" \
  *.proto

if [ ! -x "$(command -v protoc-gen-tsipc)" ]; then
    echo "Installing github.com/gabriel/grpcutil/protoc-gen-tsipc"
    go install github.com/gabriel/grpcutil/protoc-gen-tsipc
fi
echo "protoc-gen-tsipc"    
protoc \
  $protoc_include \
  --tsipc_out=. \
  *.proto
