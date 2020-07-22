#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

keycmd=${KEYS:-"keys"}
echo "- cmd: $keycmd"

infile=`mktemp /tmp/XXXXXXXXXXX`
echo "test message" > $infile
outfile=`mktemp /tmp/XXXXXXXXXXX`

# echo "list"
# kid=`keys list | head -1 | cut -d ' ' -f 1`
echo "- gen"
kid=`$keycmd generate`
echo "- gen $kid"

echo "- sign/verify (stdin) $kid"
cat "$infile" | $keycmd sign -s $kid | $keycmd verify -s $kid > $outfile
diff "$infile" "$outfile"

echo "- sign/verify (stdin, binary) $kid"
cat "$infile" | $keycmd sign -m binary -s $kid | $keycmd verify -m binary -s $kid > $outfile
diff "$infile" "$outfile"

echo "- sign (stdin, detached) $kid"
cat "$infile" | $keycmd sign -m detached -s "$kid" > "$outfile"
echo "- verify (stdin, detached)"
cat "$infile" | $keycmd verify -x "$outfile" -s $kid

echo "- sign (stdin, binary, detached) $kid"
cat "$infile" | $keycmd sign -m binary,detached -s "$kid" > "$outfile"
echo "- verify (stdin, binary, detached)"
cat "$infile" | $keycmd verify -m binary -x "$outfile" -s $kid

echo "- remove $kid"
$keycmd remove "$kid"