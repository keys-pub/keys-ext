#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd $dir

echo "auth"
eval $(keys -app Test auth -password "testpassword123")

export KEYS="keys -app Test"
./key.sh
./encrypt.sh
./sign.sh

keys -app Test uninstall
keysd -app Test -reset-keyring -force