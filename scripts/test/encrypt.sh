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
head -c 500000 </dev/urandom > "$infile"
echo "- infile: $infile"

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
cat "$infile.aenc" | $keycmd decrypt > "$infile.aorig"
diff "$infile" "$infile.aorig"

echo "- encrypt (file) $kid"
$keycmd encrypt -recipient $kid -in "$infile" -out "$infile.enc"
echo "- decrypt (file)"
$keycmd decrypt -in "$infile.enc" -out "$infile.forig"
diff "$infile" "$infile.forig"
echo "- decrypt (file, default out)"
rm "$infile"
$keycmd decrypt -in "$infile.enc"
diff "$infile" "$infile.forig"
echo "- decrypt (file, unrecognized ext)"
mv "$infile.enc" "$infile.dat"
$keycmd decrypt -in "$infile.dat"
diff "$infile.dat.dec" "$infile"

echo "- encrypt/decrypt (signcrypt, piped)"
echo "testing" > $infile
cat $infile | $keycmd encrypt -r $kid -s $kid -m signcrypt -a | $keycmd decrypt > "/tmp/test.out"
diff $infile "/tmp/test.out"
cat $infile | $keycmd encrypt -r $kid -s $kid -m signcrypt | $keycmd decrypt > "/tmp/test.out"
diff $infile "/tmp/test.out"

echo "- remove $kid"
$keycmd remove "$kid"

echo "-"
echo "- encrypt/decrypt success"
echo "-"