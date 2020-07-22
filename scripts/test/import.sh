#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

keyfile=`mktemp /tmp/XXXXXXXXXXX`

keycmd=${KEYS:-"keys"}
echo "- cmd: $keycmd"

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