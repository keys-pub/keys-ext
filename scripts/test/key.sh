#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

echo "gen"
kid=`keys generate`
echo "gen $kid"

echo "export"
keys export -kid $kid -password "testpassword123" > key.txt

echo "remove $kid"
keys remove -kid "$kid"

echo "import"
cat key.txt | keys import -password "testpassword123"

echo "remove $kid"
keys remove "$kid"

rm key.txt

