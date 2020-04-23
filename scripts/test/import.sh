#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

keyfile=`mktemp /tmp/XXXXXXXXXXX`

keycmd=${KEYS:-"keys"}
echo "cmd: $keycmd"

echo "gen"
kid=`$keycmd generate`
echo "gen $kid"

echo "export $kid"
$keycmd export -kid $kid -password testpassword123 > "$keyfile"
echo "remove $kid"
$keycmd remove "$kid"

echo "import (in)"
$keycmd import -in "$keyfile" -password testpassword123
echo "remove $kid"
$keycmd remove "$kid"

echo "import (stdin)"
cat "$keyfile" | $keycmd import -password testpassword123
echo "remove $kid"
$keycmd remove "$kid"

echo "import (stdin, kid)"
echo "$kid" | $keycmd import
echo "remove $kid"
$keycmd remove "$kid"