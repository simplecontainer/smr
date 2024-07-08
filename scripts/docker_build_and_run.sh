#!/bin/bash
cd "$(dirname "$0")"
cd ..

echo "Doing work in directory $PWD"

BRANCH="main"
BASE_DIR="$PWD"
LATEST_SMR_COMMIT="$(git rev-parse --short $BRANCH)"

cd "$BASE_DIR"

docker stop $(docker ps -q)
docker rm $(docker ps -aq)
docker image rm smr:$LATEST_SMR_COMMIT || echo "image not existing"
docker build . --file docker/Dockerfile --tag smr:$LATEST_SMR_COMMIT
docker run -v /var/run/docker.sock:/var/run/docker.sock -v /tmp:/tmp -v /home/qdnqn/testing-smr:/home/smr-agent/.ssh -p 0.0.0.0:1443:1443 --name smr-agent --dns 127.0.0.1 smr:$LATEST_SMR_COMMIT -it