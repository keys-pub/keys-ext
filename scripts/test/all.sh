#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$dir"

keys stop || true

function cleanup {
    echo "Cleaning up..."
    keys -app Test uninstall -force
}
trap cleanup EXIT

echo "- auth"
eval $(keys -app Test auth -password "testpassword123")

export KEYS="keys -app Test"
./encrypt.sh
./sign.sh
./import.sh

echo "Success!"

# TODO: Add to CI