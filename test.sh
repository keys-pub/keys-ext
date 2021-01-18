#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$dir"


(cd auth/fido2 && go test -v ./...)          
(cd auth/mock && go test -v ./...)          
(cd http/api && go test -v ./...)
(cd http/client && go test -v ./...)
(cd http/server && go test -v ./...)
(cd sdb && go test -v ./...)
(cd service && go test -v ./...)
(cd sqlchipher && go test -v ./...)
(cd vault && go test -v ./...)
# (cd wormhole && go test -v ./...)