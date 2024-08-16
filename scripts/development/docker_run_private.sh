#!/bin/bash
cd "$(dirname "$0")"
cd ../../

echo "Doing work in directory $PWD"

BASE_DIR="$PWD"
BRANCH="$(git rev-parse --abbrev-ref HEAD)"
LATEST_SMR_COMMIT="$(git rev-parse --short $BRANCH)"

mkdir -p $HOME/.smr
docker stop smr-agent
docker rm smr-agent

docker run \
       -v /var/run/docker.sock:/var/run/docker.sock \
       -v $HOME/.smr:/home/smr-agent/smr \
       -e DOMAIN=localhost \
       -e EXTERNALIP=127.0.0.1 \
       smr:$LATEST_SMR_COMMIT create

docker run \
       -v /var/run/docker.sock:/var/run/docker.sock \
       -v $HOME/.smr:/home/smr-agent/smr \
       -v $HOME/.ssh:/home/smr-agent/.ssh \
       -v /tmp:/tmp \
       -p 127.0.0.1:1443:1443 \
       --dns 127.0.0.1 \
       --name smr-agent \
       -e DOMAIN=localhost \
       -e EXTERNALIP=127.0.0.1 \
       -d smr:$LATEST_SMR_COMMIT start