#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

tmpdir=`mktemp -d 2>/dev/null || mktemp -d -t 'mytmpdir'`
cd $tmpdir

ver=${1:-""}

if [ "$ver" = "" ]; then 
  echo "Specify version to release, vx.y.z"
  exit 1
fi

if [[ ! $ver == v* ]]; then
  echo "Version should start with v"
  exit 1
fi

ver="${ver:1:${#ver}-1}"
echo "Version: $ver"

# Sync from remote to local
echo "Syncing remote apt repo..."
if [ -d "$HOME/.aptly/public/pool" ]; then 
  rm -rf /tmp/pool
  mv $HOME/.aptly/public/pool /tmp
  rm -rf $HOME/.aptly
  mkdir -p $HOME/.aptly/public
  mv /tmp/pool $HOME/.aptly/public
else
  mkdir -p $HOME/.aptly/public
fi

gsutil -m rsync -r gs://aptly.keys.pub $HOME/.aptly/public

# Get release
wget https://github.com/keys-pub/keys-ext/releases/download/v${ver}/keys_${ver}_linux_i386.deb
wget https://github.com/keys-pub/keys-ext/releases/download/v${ver}/keys_${ver}_linux_amd64.deb
# wget https://github.com/keys-pub/keys-ext/releases/download/v${ver}/keys_${ver}_linux_armv6.deb

echo "Create apt repo"
aptly repo create -distribution=current -component=main keys-release
# aptly repo show keys-release
aptly repo add keys-release . 
aptly snapshot create keys-${ver} from repo keys-release
# aptly publish switch -gpg-key=B1A671AD current keys-${ver}
echo "Publish snapshot"
aptly publish snapshot -distribution=current -gpg-key=B1A671AD -gpg-provider=gpg2 keys-${ver}

echo "Publishing to remote..."
gsutil -m rsync -r $HOME/.aptly/public gs://aptly.keys.pub

