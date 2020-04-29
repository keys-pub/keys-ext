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
tmpdir=`mktemp -d 2>/dev/null || mktemp -d -t 'mytmpdir'`
echo "Building FIDO2 plugin ($tmpdir/fido2.so)"
cd $dir/../service/fido2
go build -buildmode=plugin -o "$tmpdir/fido2.so"
sh "$dir/codesign.sh" "$tmpdir/fido2.so"
mv "$tmpdir/fido2.so" ~/go/bin 