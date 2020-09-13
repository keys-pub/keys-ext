#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$dir"

scripts="$dir/../scripts"

ts=`date +%s`
date=`date -u +"%Y-%m-%dT%H:%M:%SZ"`
ver="0.0.0-dev-$ts"

DEBUG=1 VERSION=$ver DATE=$date "$scripts/gobuild.sh" keysd "$dir/../service/keysd"
DEBUG=1 VERSION=$ver DATE=$date "$scripts/gobuild.sh" keys  "$dir/../service/keys"

