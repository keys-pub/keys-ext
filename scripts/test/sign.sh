#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

infile=`mktemp /tmp/XXXXXXXXXXX`
head -c 500000 </dev/urandom > "$infile"
echo "infile: $infile"

sigfile=`mktemp /tmp/XXXXXXXXXXX`
outfile=`mktemp /tmp/XXXXXXXXXXX`

keycmd=${KEYS:-"keys"}
echo "cmd: $keycmd"

# echo "list"
# kid=`keys list | head -1 | cut -d ' ' -f 1`
echo "gen"
kid=`$keycmd generate`
echo "gen $kid"

echo "sign $kid"
$keycmd sign -s "$kid" -in "$infile" -out "$sigfile"
echo "verify"
$keycmd verify -s $kid -in "$sigfile" -out "$outfile"
diff "$infile" "$outfile"

echo "sign (stdin/stdout) $kid"
cat "$infile" | $keycmd sign -stdin -stdout -s "$kid" > "$sigfile"
echo "verify (stdin/stdout)"
cat "$sigfile" | $keycmd verify -stdin -stdout -s $kid > "$outfile"
diff "$infile" "$outfile"

echo "remove $kid"
$keycmd remove "$kid"