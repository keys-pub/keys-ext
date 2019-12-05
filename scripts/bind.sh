#!/usr/bin/env bash

set -e -o pipefail # Fail on error
# Not using -u for unbound variables

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$dir"

if [ "$GOPATH" = "" ]; then
  echo "Using default GOPATH: $GOPATH"
  export GOPATH="$HOME/go"
fi

echo "Building Keys.framework..."

(cd ../js/ios/ && "$GOPATH/bin/gomobile" bind -target ios github.com/gabriel/keysbind)
