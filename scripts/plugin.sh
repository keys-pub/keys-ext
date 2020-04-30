#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd $dir

dest=$1
version=$2
date=$3
path=$4

# TODO: Pass in .Os from goreleaser when that works
if [[ "$path" = *".exe" ]]; then
    echo "Skipping windows"
    exit 0
fi

BUILD_ONLY=1 DEBUG=1 VERSION=$version DATE=$date DEST=$dest BUILD_FLAGS="-buildmode=plugin" ./gobuild.sh fido2.so "$dir/../service/fido2"