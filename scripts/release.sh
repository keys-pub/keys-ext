#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

# Checkout to tmpdir
tmpdir=`mktemp -d 2>/dev/null || mktemp -d -t 'mytmpdir'`
echo "$tmpdir"
cd "$tmpdir"
git clone https://github.com/keys-pub/keys-ext
cd keys-ext

# keys, keysd
cd service
goreleaser

# fido2.so
cd ../auth/rpc/plugin
goreleaser

# Cleanup
cd "$dir"
rm -rf "$tmpdir"