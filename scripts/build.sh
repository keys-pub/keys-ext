#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd $dir

ts=`date +%s`
date=`date -u +"%Y-%m-%dT%H:%M:%SZ"`
ver="0.0.0-dev-$ts"

os=`go env GOOS`

BUILD_ONLY=1 DEBUG=1 VERSION=$ver DATE=$date ./gobuild.sh keysd "$dir/../service/keysd"
BUILD_ONLY=1 DEBUG=1 VERSION=$ver DATE=$date ./gobuild.sh keys  "$dir/../service/keys"
if [ ! "$os" = "windows" ]; then
    BUILD_ONLY=1 DEBUG=1 VERSION=$ver DATE=$date BUILD_FLAGS="-buildmode=plugin" ./gobuild.sh fido2.so "$dir/../service/fido2"
fi

