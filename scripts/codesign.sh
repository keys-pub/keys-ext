#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

code_sign_identity="Developer ID Application: Gabriel Handford (U2622K69A6)"
codesign --verbose --sign "$code_sign_identity" "$1"
