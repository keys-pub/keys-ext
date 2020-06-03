#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

# Checkout to tmpdir
tmpdir=`mktemp -d 2>/dev/null || mktemp -d -t 'mytmpdir'`
echo "$tmpdir"
cd "$tmpdir"
git clone https://github.com/keys-pub/keys-ext
cd keys-ext

ver=`git describe --abbrev=0 --tags`
echo "Latest tag: $ver"
git checkout $ver

cd service
goreleaser --config=.goreleaser.linux.yml

cd ../auth/rpc/plugin
goreleaser --config=.goreleaser.linux.yml

$dir/aptly.sh $ver

# Cleanup
cd $dir
rm -rf "$tmpdir"