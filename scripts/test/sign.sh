#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

infile=`mktemp /tmp/XXXXXXXXXXX`
head -c 500000 </dev/urandom > "$infile"
echo "infile: $infile"

keycmd=${KEYS:-"keys"}
echo "cmd: $keycmd"

# echo "list"
# kid=`keys list | head -1 | cut -d ' ' -f 1`
echo "gen"
kid=`$keycmd generate`
echo "gen $kid"

echo "sign (stdin) $kid"
cat "$infile" | $keycmd sign -s "$kid" > "$infile.sig"
echo "verify (stdin)"
cat "$infile.sig" | $keycmd verify -s $kid > "$infile.out"
diff "$infile" "$infile.out"

echo "sign (stdin, armor) $kid"
cat "$infile" | $keycmd sign -a -s "$kid" > "$infile.asig"
echo "verify (stdin)"
cat "$infile.asig" | $keycmd verify -a -s $kid > "$infile.aout"
diff "$infile" "$infile.aout"

echo "sign (stdin, detached) $kid"
cat "$infile" | $keycmd sign -d -s "$kid" > "$infile.asig"
echo "verify (stdin, detached)"
cat "$infile" | $keycmd verify -x "$infile.asig" -s $kid

echo "sign (stdin, armor, detached) $kid"
cat "$infile" | $keycmd sign -a -d -s "$kid" > "$infile.asig"
echo "verify (stdin, armor, detached)"
cat "$infile" | $keycmd verify -a -x "$infile.asig" -s $kid

echo "remove $kid"
$keycmd remove "$kid"