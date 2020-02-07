#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

skip_tag=${SKIP_TAG:-""}
skip_linux=${SKIP_LINUX:-""}

# Checkout to tmpdir
tmpdir=`mktemp -d 2>/dev/null || mktemp -d -t 'mytmpdir'`
echo "$tmpdir"
cd "$tmpdir"
git clone https://github.com/keys-pub/keysd
cd keysd

# Tag
if [ ! "$skip_tag" = "1" ]; then
    $dir/tag.sh `pwd`
fi

# Linux
if [ ! "$skip_linux" = "1" ]; then
    goreleaser --config=.goreleaser.linux.yml --rm-dist
    ver=`git describe --abbrev=0 --tags`
    $dir/aptly.sh $ver
fi

# Other platforms
goreleaser --rm-dist

# Cleanup
cd $dir
rm -rf "$tmpdir"