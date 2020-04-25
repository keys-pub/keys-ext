#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

cd "$1"

go="-I=`go list -f '{{ .Dir }}' -m github.com/golang/protobuf`"
gogo="-I=`go list -f '{{ .Dir }}' -m github.com/gogo/protobuf`"
#fido2="-I=`realpath ../fido2`"

protoc_include="-I ./ $go $gogo"
#protoc_include="-I ./ $go $gogo $fido2"

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

echo "protoc $protoc_include --gogo_out=plugins=grpc:. *.proto"
protoc $protoc_include --gogo_out=plugins=grpc:. *.proto

# tstypes
if [ ! -x "$(command -v protoc-gen-tstypes)" ]; then
    echo "Installing github.com/tmc/grpcutil/protoc-gen-tstypes"
    go install github.com/tmc/grpcutil/protoc-gen-tstypes
fi
protoc $protoc_include --tstypes_out=. --tstypes_opt=declare_namespace=false,outpattern="{{.BaseName}}.d.ts" *.proto

# tsipc
if [ ! -x "$(command -v protoc-gen-tsipc)" ]; then
    echo "Installing github.com/gabriel/grpcutil/protoc-gen-tsipc"
    go install github.com/gabriel/grpcutil/protoc-gen-tsipc
fi
echo "protoc $protoc_include --tsipc_out=. *.proto"    
protoc $protoc_include --tsipc_out=. *.proto

# CLI
# go get github.com/gabriel/grpcutil/protoc-gen-gocli
# protoc $protoc_include --gocli_out=. --gocli_opt=logtostderr=true,v=0 keys.proto
# sed -i '' 's/package proto/package service/g' keys.go
# mv keys.go keys.cli.go

# Mocks
# mockgen -destination=mock/keys.go github.com/gabriel/keys/service/proto ChatClient

