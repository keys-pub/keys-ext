#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd $dir

ts=`date +%s`
date=`date -u +"%Y-%m-%dT%H:%M:%SZ"`
ver="0.0.0-dev-$ts"

BUILD_ONLY=1 DEBUG=1 VERSION=$ver DATE=$date ./gobuild.sh keysd "$dir/../service/keysd"
BUILD_ONLY=1 DEBUG=1 VERSION=$ver DATE=$date ./gobuild.sh keys  "$dir/../service/keys"

# FIDO2
echo "Building FIDO2 plugin"
cd ../service/fido2
go build -buildmode=plugin -o fido2.so
mv fido2.so ~/go/bin 