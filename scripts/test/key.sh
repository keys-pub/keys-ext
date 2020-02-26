#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

keycmd=${KEYS:-"keys"}

echo "gen"
kid=`${keycmd} generate`
echo "gen $kid"

echo "export"
$keycmd export -kid $kid -password "testpassword123" > key.txt

echo "remove $kid"
$keycmd remove -kid "$kid"

echo "import"
cat key.txt | $keycmd import -password "testpassword123"
rm key.txt

echo "remove $kid"
$keycmd remove "$kid"


