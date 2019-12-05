#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

tmpfile=`mktemp /tmp/XXXXXXXXXXX`
tmpfile2=`mktemp /tmp/XXXXXXXXXXX`
tmpfile3=`mktemp /tmp/XXXXXXXXXXX`

head -c 500000 </dev/urandom > "$tmpfile"

sigfile=`mktemp /tmp/XXXXXXXXXXX`
sigfile2=`mktemp /tmp/XXXXXXXXXXX`

kid=`keys | head -1 | cut -d ' ' -f 1`

echo "sign"
keys sign -in "$tmpfile" -out "$sigfile"
echo "verify"
keys verify -kid $kid -in "$sigfile" -out "$tmpfile2"
diff "$tmpfile" "$tmpfile2"

echo "sign"
cat "$tmpfile2" | keys sign > "$sigfile2"
echo "verify"
cat "$sigfile2" | keys verify -kid $kid > "$tmpfile3"
diff "$tmpfile" "$tmpfile3"


