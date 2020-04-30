#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

bin=$1

# TODO: Pass in .Os from goreleaser when that works
if [[ "$bin" = *".exe" ]]; then
    echo "Skipping windows"
    exit 0
fi

code_sign_identity="Developer ID Application: Gabriel Handford (U2622K69A6)"
echo "Signing: $code_sign_identity $bin"
codesign --verbose --sign "$code_sign_identity" "$bin"
