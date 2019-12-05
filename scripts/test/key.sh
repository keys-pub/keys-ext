#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

echo "gen"
kid=`keys generate -publish`
echo "gen $kid"
echo "backup"
seed=`keys backup -kid $kid`
echo "backup $seed"
echo "remove"
keys remove -seed-phrase "$seed" -kid "$kid"
echo "pull"
keys pull -kid "$kid"


echo "gen"
kid=`keys generate`
echo "gen $kid"
echo "push"
keys push -kid "$kid"
echo "remove"
keys remove -seed-phrase "`keys backup -kid $kid`" -kid "$kid"
echo "pull"
keys pull -kid "$kid"
