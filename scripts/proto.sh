#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$dir"
cd ../service

keysapp="$dir/../../app"

inc1="-I=`go list -f '{{ .Dir }}' -m github.com/golang/protobuf`"
inc2="-I=`go list -f '{{ .Dir }}' -m github.com/gogo/protobuf`"

protoc_include="-I ./ $inc1 $inc2"

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

if [ -d "$keysapp" ]; then
    cp keys.proto "$keysapp/src/renderer/rpc/keys.proto"

    # tstypes
    if [ ! -x "$(command -v protoc-gen-tstypes)" ]; then
        echo "Installing github.com/tmc/grpcutil/protoc-gen-tstypes"
        go install github.com/tmc/grpcutil/protoc-gen-tstypes
    fi
    protoc $protoc_include --tstypes_out=. --tstypes_opt=declare_namespace=false keys.proto
    mv service.keys.d.ts "$keysapp/src/renderer/rpc/types.ts"

    # tsrpc (redux)
    if [ ! -x "$(command -v protoc-gen-tsrpc)" ]; then
        echo "Installing github.com/gabriel/grpcutil/protoc-gen-tsrpc"
        go install github.com/gabriel/grpcutil/protoc-gen-tsrpc
    fi
    echo "protoc $protoc_include --tsrpc_out=. keys.proto"    
    protoc $protoc_include --tsrpc_out=. keys.proto
    mv keys.ts "$keysapp/src/renderer/rpc/rpc.ts"
fi

# CLI
# go get github.com/gabriel/grpcutil/protoc-gen-gocli
# protoc $protoc_include --gocli_out=. --gocli_opt=logtostderr=true,v=0 keys.proto
# sed -i '' 's/package proto/package service/g' keys.go
# mv keys.go keys.cli.go

# Mocks
# mockgen -destination=mock/keys.go github.com/gabriel/keys/service/proto ChatClient

