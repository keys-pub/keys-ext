#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

code_sign_identity="Developer ID Application: Gabriel Handford (U2622K69A6)"
codesign --verbose --sign "$code_sign_identity" "$dir/../$1"
