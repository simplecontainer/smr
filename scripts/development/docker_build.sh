#!/bin/bash
cd "$(dirname "$0")"
cd ../../

echo "Doing work in directory $PWD"

BASE_DIR="$PWD"
BRANCH="$(git rev-parse --abbrev-ref HEAD)"
LATEST_SMR_COMMIT="$(git rev-parse --short $BRANCH)"

#docker build . --file docker/Dockerfile --no-cache --tag smr:$LATEST_SMR_COMMIT
docker buildx build --file docker/Dockerfile --tag smr:$LATEST_SMR_COMMIT --platform linux/amd64,linux/arm64 .
