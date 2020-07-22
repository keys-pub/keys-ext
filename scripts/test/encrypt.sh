#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

infile=`mktemp /tmp/XXXXXXXXXXX`
head -c 500000 </dev/urandom > "$infile"
echo "- infile: $infile"

keycmd=${KEYS:-"keys"}
echo "- cmd: $keycmd"

echo "- gen"
kid=`$keycmd generate`
echo "- gen $kid"

echo "- encrypt (stdin) $kid"
cat "$infile" | $keycmd encrypt -recipient $kid > "$infile.enc"
echo "- decrypt (stdin)"
cat "$infile.enc" | $keycmd decrypt > "$infile.orig"
diff "$infile" "$infile.orig"

echo "- encrypt (stdin, armor) $kid"
cat "$infile" | $keycmd encrypt -a -recipient $kid > "$infile.aenc"
echo "- decrypt (stdin)"
cat "$infile.aenc" | $keycmd decrypt -a > "$infile.aorig"
diff "$infile" "$infile.aorig"

echo "- remove $kid"
$keycmd remove "$kid"