#!/bin/bash
cd "$(dirname "$0")"
cd ../../

echo "Doing work in directory $PWD"

BASE_DIR="$PWD"
BRANCH="$(git rev-parse --abbrev-ref HEAD)"
LATEST_SMR_COMMIT="$(git rev-parse --short $BRANCH)"

docker build . --file docker/Dockerfile --no-cache --build-arg TARGETOS=linux --build-arg TARGETARCH=amd64 --tag smr:$LATEST_SMR_COMMIT
