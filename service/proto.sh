#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$dir"

"$dir/../scripts/proto3.sh" "$dir"

tsclient="$dir/../../tsclient"
if [ -d "$tsclient" ]; then
    echo "Copying keys.proto to $tsclient"
    cp keys.proto "$tsclient/proto/keys.proto"

    cd "$tsclient"    
    yarn build
fi
