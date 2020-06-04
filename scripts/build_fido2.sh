#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd $dir

ts=`date +%s`
date=`date -u +"%Y-%m-%dT%H:%M:%SZ"`
ver="0.0.0-dev-$ts"

BUILD_ONLY=1 DEBUG=1 VERSION=$ver DATE=$date BUILD_FLAGS="-buildmode=plugin -o fido2.so" ./gobuild.sh fido2.so "$dir/../auth/rpc/plugin"


