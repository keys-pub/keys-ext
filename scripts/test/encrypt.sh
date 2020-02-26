#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

tmpfile=`mktemp /tmp/XXXXXXXXXXX`
tmpfile2=`mktemp /tmp/XXXXXXXXXXX`
tmpfile3=`mktemp /tmp/XXXXXXXXXXX`

head -c 500000 </dev/urandom > "$tmpfile"

encfile=`mktemp /tmp/XXXXXXXXXXX`
encfile2=`mktemp /tmp/XXXXXXXXXXX`

keycmd=${KEYS:-"keys"}

echo "gen"
kid=`$keycmd generate`
echo "gen $kid"

echo "encrypt $kid"
$keycmd encrypt -recipient $kid -in "$tmpfile" -out "$encfile"
echo "decrypt"
$keycmd decrypt -in "$encfile" -out "$tmpfile2"
diff "$tmpfile2" "$tmpfile2"

echo "encrypt $kid"
cat "$tmpfile2" | $keycmd encrypt -recipient $kid > "$encfile2"
echo "decrypt"
cat "$encfile2" | $keycmd decrypt > "$tmpfile3"
diff "$tmpfile2" "$tmpfile3"

echo "remove $kid"
$keycmd remove "$kid"