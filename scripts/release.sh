#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

skip_tag=${SKIP_TAG:-""}

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

cd service
goreleaser --rm-dist

# Cleanup
cd $dir
rm -rf "$tmpdir"