#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

outfile=`mktemp /tmp/XXXXXXXXXXX`

keycmd=${KEYS:-"keys"}
echo "cmd: $keycmd"

echo "gen"
kid=`$keycmd generate`
echo "gen $kid"

echo "export $kid"
$keycmd export -kid $kid -password testpassword123 > "$outfile"
echo "import (in)"
$keycmd import -in "$outfile" -password testpassword123

echo "import (stdin)"
cat "$outfile" | $keycmd import -stdin -password testpassword123

echo "remove $kid"
$keycmd remove "$kid"