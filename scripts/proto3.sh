#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

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

tsclient="$dir/../../tsclient"
if [ -d "$tsclient" ]; then
    echo "Copying proto to $tsclient"
    cp *.proto "$tsclient/proto/"
    echo "Building tsclient..."
    (cd "$tsclient" && yarn build)
fi
