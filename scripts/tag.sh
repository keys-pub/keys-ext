#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=${1:-""}

if [ "$dir" = "" ]; then 
  echo "Specify repo directory"
  exit 1
fi

# $1 - semver string
# $2 - level to incr {release,minor,major} - release by default
function incr_semver() { 
    IFS='.' read -ra ver <<< "$1"
    [[ "${#ver[@]}" -ne 3 ]] && echo "Invalid semver string" && return 1
    [[ "$#" -eq 1 ]] && level='release' || level=$2

    release=${ver[2]}
    minor=${ver[1]}
    major=${ver[0]}

    case $level in
        release)
            release=$((release+1))
        ;;
        minor)
            release=0
            minor=$((minor+1))
        ;;
        major)
            release=0
            minor=0
            major=$((major+1))
        ;;
        *)
            echo "Invalid level passed"
            return 2
    esac
    echo "$major.$minor.$release"
}

cd $dir
ver=`git describe --abbrev=0 --tags`
echo "Version: $ver"

next=$(incr_semver $ver "release")
echo "Next: $next"

# echo " "
# echo "To tag, run these commands:"
# echo " "
# echo "git tag -a $next -m $next"
# echo "git push --tags"
# echo " "

if [[ ! $next == v* ]]; then
  echo "Tag should start with v"
  exit 1
fi

echo "git tag -a $next -m $next"
git tag -a $next -m $next