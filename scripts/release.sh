#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

# Checkout to tmpdir
tmpdir=`mktemp -d 2>/dev/null || mktemp -d -t 'mytmpdir'`
echo "$tmpdir"
cd "$tmpdir"
git clone https://github.com/keys-pub/keysd
cd keysd

# Tag
$dir/tag.sh `pwd`

# Linux
goreleaser --config=.goreleaser.linux.yml --rm-dist
ver=`git describe --abbrev=0 --tags`
$dir/aptly.sh $ver

# Other platforms
goreleaser --rm-dist

# Cleanup
cd $dir
rm -rf "$tmpdir"