#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

tmpfile=`mktemp /tmp/XXXXXXXXXXX`
tmpfile2=`mktemp /tmp/XXXXXXXXXXX`
tmpfile3=`mktemp /tmp/XXXXXXXXXXX`

head -c 500000 </dev/urandom > "$tmpfile"

encfile=`mktemp /tmp/XXXXXXXXXXX`
encfile2=`mktemp /tmp/XXXXXXXXXXX`

# echo "list"
# kid=`keys list | head -1 | cut -d ' ' -f 1`
echo "gen"
kid=`keys generate`
echo "gen $kid"

echo "encrypt $kid"
keys encrypt -recipient $kid -in "$tmpfile" -out "$encfile"
echo "decrypt"
keys decrypt -in "$encfile" -out "$tmpfile2"
diff "$tmpfile2" "$tmpfile2"

echo "encrypt $kid"
cat "$tmpfile2" | keys encrypt -recipient $kid > "$encfile2"
echo "decrypt"
cat "$encfile2" | keys decrypt > "$tmpfile3"
diff "$tmpfile2" "$tmpfile3"

echo "remove $kid"
keys remove "$kid"