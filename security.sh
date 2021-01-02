#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$dir"

go get github.com/securego/gosec/cmd/gosec; 
(cd auth/fido2 && gosec ./...)
(cd auth/mock && gosec ./...) 
(cd auth/rpc && gosec ./...) 
(cd firestore && gosec ./...) 
(cd http/api && gosec ./...) 
(cd http/client && gosec ./...)
(cd http/server && gosec ./...) 
(cd sdb && gosec ./...)
(cd service && gosec ./...)
(cd vault && gosec ./...)
(cd wormhole && gosec .)
(cd ws/api && gosec .)
(cd ws/client && gosec .) 
(cd ws/server && gosec .)