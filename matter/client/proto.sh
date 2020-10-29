#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$dir"

# protoc-gen-go
if [ ! -x "$(command -v protoc-gen-go)" ]; then
    echo "Installing google.golang.org/protobuf/cmd/protoc-gen-go"
    go install google.golang.org/protobuf/cmd/protoc-gen-go
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