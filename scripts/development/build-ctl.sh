#!/bin/bash
cd "$(dirname "$0")" || exit 1
cd ../../

echo "Doing work in directory $PWD"

BASE_DIR="$PWD"

cd "$BASE_DIR/cmd/smrctl" || exit 1

echo "***********************************************"
echo "$(pwd)"
echo "***********************************************"

CGO_ENABLED=0 go build -ldflags '-s -w' || exit 1

mkdir $BASE_DIR/smrctl-linux-amd64
cp -f smrctl $BASE_DIR/smrctl-linux-amd64/smrctl

cd "$BASE_DIR" || exit 1