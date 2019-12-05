#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

echo "gen"
kid=`keys generate -publish`
echo "gen $kid"
echo "remove $kid"
keys remove -kid "$kid" -seed-phrase "`keys backup -kid $kid`" 
