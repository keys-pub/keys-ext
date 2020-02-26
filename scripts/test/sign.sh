#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

tmpfile=`mktemp /tmp/XXXXXXXXXXX`
tmpfile2=`mktemp /tmp/XXXXXXXXXXX`
tmpfile3=`mktemp /tmp/XXXXXXXXXXX`

head -c 500000 </dev/urandom > "$tmpfile"

sigfile=`mktemp /tmp/XXXXXXXXXXX`
sigfile2=`mktemp /tmp/XXXXXXXXXXX`

keycmd=${KEYS:-"keys"}

# echo "list"
# kid=`keys list | head -1 | cut -d ' ' -f 1`
echo "gen"
kid=`$keycmd generate`
echo "gen $kid"

echo "sign $kid"
$keycmd sign -s "$kid" -in "$tmpfile" -out "$sigfile"
echo "verify"
$keycmd verify -s $kid -in "$sigfile" -out "$tmpfile2"
diff "$tmpfile" "$tmpfile2"

echo "sign $kid"
cat "$tmpfile2" | $keycmd sign -s "$kid" > "$sigfile2"
echo "verify"
cat "$sigfile2" | $keycmd verify -s $kid > "$tmpfile3"
diff "$tmpfile" "$tmpfile3"

echo "remove $kid"
$keycmd remove "$kid"