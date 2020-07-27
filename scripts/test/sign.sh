#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

function cleanup {
    echo "Cleaning up..."
    keys -app Test uninstall -force
}
trap cleanup EXIT

keycmd=${KEYS:-"keys -app Test"}
echo "- cmd: $keycmd"
keys stop || true
echo "- auth"
eval $(keys -app Test auth -password "testpassword123")

infile=`mktemp /tmp/XXXXXXXXXXX`
echo "test message" > $infile
outfile=`mktemp /tmp/XXXXXXXXXXX`

# echo "list"
# kid=`keys list | head -1 | cut -d ' ' -f 1`
echo "- gen"
kid=`$keycmd generate`
echo "- gen $kid"

echo "- sign/verify (stdin) $kid"
cat "$infile" | $keycmd sign -s $kid | $keycmd verify > $outfile
diff "$infile" "$outfile"

echo "- sign/verify (stdin, binary, expected) $kid"
cat "$infile" | $keycmd sign -m binary -s $kid | $keycmd verify -s $kid > $outfile
diff "$infile" "$outfile"

echo "- sign (stdin, detached) $kid"
cat "$infile" | $keycmd sign -m detached -s $kid > "$outfile"
echo "- verify (stdin, detached)"
cat "$infile" | $keycmd verify -x "$outfile"
echo "- verify (stdin, detached, expected)"
cat "$infile" | $keycmd verify -x "$outfile" -s $kid

echo "- sign (stdin, binary, detached) $kid"
cat "$infile" | $keycmd sign -m binary,detached -s $kid > "$outfile"
echo "- verify (stdin, binary, detached)"
cat "$infile" | $keycmd verify -x "$outfile"
echo "- verify (stdin, binary, detached, expected)"
cat "$infile" | $keycmd verify -x "$outfile" -s "$kid"

echo "- remove $kid"
$keycmd remove "$kid"

echo "- sign/verify success"