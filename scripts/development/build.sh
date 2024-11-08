#!/bin/bash
cd "$(dirname "$0")"
cd ../../

echo "Doing work in directory $PWD"

BASE_DIR="$PWD"
BRANCH="$(git rev-parse --abbrev-ref HEAD)"
LATEST_SMR_COMMIT="$(git rev-parse --short $BRANCH)"

cd "$BASE_DIR"

echo "***********************************************"
echo "$BASE_DIR/$DIRNAME"
echo "***********************************************"

CGO_ENABLED=0 go build -ldflags '-s -w' || exit 1
mv smr smr-linux-amd64

cd "$BASE_DIR"