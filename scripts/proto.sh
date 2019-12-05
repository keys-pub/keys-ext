#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$dir"
cd ../service

keysapp="$dir/../../keys-app"

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
    cp keys.proto "$keysapp/app/rpc"

    # Flow
    # For more debugging output, change verbose level, for example, v=0 to v=1
    
    if [ ! -x "$(command -v protoc-gen-flowtypes)" ]; then
        echo "Installing github.com/gabriel/grpcutil/protoc-gen-flowtypes"
        go install github.com/gabriel/grpcutil/protoc-gen-flowtypes
    fi
    protoc $protoc_include --flowtypes_out=. --flowtypes_opt=logtostderr=true,v=0,enum_zeros=true keys.proto
    mv keys.js "$keysapp/app/rpc/types.js"

    # JSRPC (redux)
    if [ ! -x "$(command -v protoc-gen-jsrpc)" ]; then
        echo "Installing github.com/gabriel/grpcutil/protoc-gen-jsrpc"
        go install github.com/gabriel/grpcutil/protoc-gen-jsrpc
    fi
    echo "protoc $protoc_include --jsrpc_out=. --jsrpc_opt=logtostderr=true,v=0 keys.proto"    
    protoc $protoc_include --jsrpc_out=. --jsrpc_opt=logtostderr=true,v=0 keys.proto
    mv keys.js "$keysapp/app/rpc/rpc.js"
fi

# CLI
# go get github.com/gabriel/grpcutil/protoc-gen-gocli
# protoc $protoc_include --gocli_out=. --gocli_opt=logtostderr=true,v=0 keys.proto
# sed -i '' 's/package proto/package service/g' keys.go
# mv keys.go keys.cli.go

# Mocks
# mockgen -destination=mock/keys.go github.com/gabriel/keys/service/proto ChatClient

