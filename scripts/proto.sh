#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

cd "$1"

go="-I=`go list -f '{{ .Dir }}' -m github.com/golang/protobuf`"
gogo="-I=`go list -f '{{ .Dir }}' -m github.com/gogo/protobuf`"

protoc_include="-I ./ $go $gogo"

if [ ! -x "$(command -v protoc-gen-go)" ]; then
    echo "Installing github.com/golang/protobuf/protoc-gen-go"
    go install github.com/golang/protobuf/protoc-gen-go
fi 

if [ ! -x "$(command -v protoc-gen-gogo)" ]; then
    echo "Installing github.com/gogo/protobuf/protoc-gen-gogo"
    go install github.com/gogo/protobuf/protoc-gen-gogo
    # go install github.com/gogo/protobuf/protoc-gen-gogoslick
    # Using gogo instead of go_out
fi 

echo "gogo"
protoc $protoc_include --gogo_out=plugins=grpc:. *.proto

# tstypes
if [ ! -x "$(command -v protoc-gen-tstypes)" ]; then
    echo "Installing github.com/tmc/grpcutil/protoc-gen-tstypes"
    go install github.com/tmc/grpcutil/protoc-gen-tstypes
fi
echo "tstypes"    
protoc $protoc_include --tstypes_out=. --tstypes_opt=declare_namespace=false,outpattern="{{.BaseName}}.d.ts" *.proto

# tsipc
if [ ! -x "$(command -v protoc-gen-tsipc)" ]; then
    echo "Installing github.com/gabriel/grpcutil/protoc-gen-tsipc"
    go install github.com/gabriel/grpcutil/protoc-gen-tsipc
fi
echo "tsipc"    
protoc $protoc_include --tsipc_out=. *.proto

