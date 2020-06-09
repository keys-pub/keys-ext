#!/usr/bin/env bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd $dir

TEST_GSUTIL=1 go test -run ^TestGSUtil$ -v
TEST_AWSS3=1 go test -run ^TestAWSS3$ -v
TEST_GIT=1 go test -run ^TestGit$ -v
TEST_RCLONE=1 go test -run ^TestRClone$ -v