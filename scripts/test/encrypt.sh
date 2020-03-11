#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

infile=`mktemp /tmp/XXXXXXXXXXX`
outfile=`mktemp /tmp/XXXXXXXXXXX`

head -c 500000 </dev/urandom > "$infile"
echo "infile: $infile"

encfile=`mktemp /tmp/XXXXXXXXXXX`
encfile2=`mktemp /tmp/XXXXXXXXXXX`

keycmd=${KEYS:-"keys"}
echo "cmd: $keycmd"

echo "gen"
kid=`$keycmd generate`
echo "gen $kid"

echo "encrypt $kid"
$keycmd encrypt -recipient $kid -in "$infile" -out "$encfile"
echo "decrypt"
$keycmd decrypt -in "$encfile" -out "$outfile"
diff "$infile" "$outfile"

echo "encrypt (armor) $kid"
$keycmd encrypt -armor -recipient $kid -in "$infile" -out "$encfile"
echo "decrypt (armor)"
$keycmd decrypt -armor -in "$encfile" -out "$outfile"
diff "$infile" "$outfile"

echo "encrypt (stdin/stdout) $kid"
cat "$infile" | $keycmd encrypt -stdin -stdout -recipient $kid > "$encfile"
echo "decrypt (stdin/stdout) $encfile"
cat "$encfile" | $keycmd decrypt -stdin -stdout > "$outfile"
diff "$infile" "$outfile"

echo "decrypt (stdin/out)"
cat "$encfile" | $keycmd decrypt -stdin -out "$outfile"
diff "$infile" "$outfile"

echo "decrypt (in/stdout)"
$keycmd decrypt -in "$encfile" -stdout > "$outfile"
diff "$infile" "$outfile"

echo "remove $kid"
$keycmd remove "$kid"