#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$dir"

PATH=$PATH:$HOME/bin

(cd auth/fido2 && golangci-lint run --timeout 10m)
(cd auth/mock && golangci-lint run --timeout 10m)
(cd auth/rpc && golangci-lint run --timeout 10m)
(cd firestore && golangci-lint run --timeout 10m)
(cd http/api && golangci-lint run --timeout 10m)
(cd http/client && golangci-lint run --timeout 10m)
(cd http/server && golangci-lint run --timeout 10m)
(cd sdb && golangci-lint run --timeout 10m)
(cd service && golangci-lint run --timeout 10m)
(cd sqlcipher && golangci-lint run --timeout 10m)
(cd vault && golangci-lint run --timeout 10m)
(cd wormhole && golangci-lint run --timeout 10m)
(cd ws/api && golangci-lint run --timeout 10m)
(cd ws/client && golangci-lint run --timeout 10m)
(cd ws/server && golangci-lint run --timeout 10m)