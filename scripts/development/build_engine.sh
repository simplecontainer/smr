#!/bin/bash
cd "$(dirname "$0")" || exit 1
cd ../../

echo "Doing work in directory $PWD"

BASE_DIR="$PWD"

cd "$BASE_DIR/cmd/smr" || exit 1

echo "***********************************************"
echo "$(pwd)"
echo "***********************************************"

CGO_ENABLED=0 go build -ldflags '-s -w' || exit 1

mkdir $BASE_DIR/smr-linux-amd64
mv smr $BASE_DIR/smr-linux-amd64/smr

cd "$BASE_DIR" || exit 1