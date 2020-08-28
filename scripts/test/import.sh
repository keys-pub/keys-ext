#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

function cleanup {
    echo "Cleaning up..."
    keys -app Test uninstall -force
}
trap cleanup EXIT

keycmd=${KEYS:-"keys -app Test"}
echo "- cmd: $keycmd"
keys stop || true
echo "- auth"
eval $(keys -app Test auth -password "testpassword123")

keyfile=`mktemp /tmp/XXXXXXXXXXX`

echo "- gen"
kid=`$keycmd generate`
echo "- gen $kid"

echo "- export"
$keycmd export -kid $kid -password testpassword123 > "$keyfile"
echo "- remove $kid"
$keycmd remove "$kid"

echo "- import (in)"
$keycmd import -in "$keyfile" -password testpassword123
echo "- remove $kid"
$keycmd remove "$kid"

echo "- gen"
kid=`$keycmd generate`
echo "- gen $kid"

echo "- import (stdin)"
cat "$keyfile" | $keycmd import -password testpassword123
echo "- remove $kid"
$keycmd remove "$kid"

echo "- import (stdin, kid)"
echo "$kid" | $keycmd import
echo "- remove $kid"
$keycmd remove "$kid"

echo "- gen"
kid=`$keycmd generate`
echo "- gen $kid"

echo "- export (no password)"
$keycmd export -kid $kid -no-password > "$keyfile"
echo "- remove $kid"
$keycmd remove "$kid"

echo "- import (no password)"
cat "$keyfile" | $keycmd import -no-password
echo "- remove $kid"
$keycmd remove "$kid"

echo "-"
echo "- import/export success"
echo "-"