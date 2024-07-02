#!/bin/bash
cd "$(dirname "$0")"
cd ..

echo "Doing work in directory $PWD"

BRANCH="main"
BASE_DIR="$PWD"
LATEST_SMR_COMMIT="$(git rev-parse --short $BRANCH)"

cd "$BASE_DIR"
go build -ldflags "-s -w"

for dir in implementations/*/
do
    DIR=${dir%*/}
    DIRNAME="${DIR##*/}"

    echo "***********************************************"
    echo "$BASE_DIR/../implementations/$DIRNAME"
    echo "***********************************************"

    cd "$BASE_DIR/implementations/$DIRNAME"
    go build -ldflags "-s -w" --buildmode=plugin
done

cd "$BASE_DIR"

#for dir in operators/*/
#do
#    DIR=${dir%*/}
#    DIRNAME="${DIR##*/}"
#
#    cd "$BASE_DIR/operators/$DIRNAME"
#    go build -ldflags "-s -w" --buildmode=plugin
#done

docker stop $(docker ps -q)
docker rm $(docker ps -aq)
docker image rm smr:$LATEST_SMR_COMMIT || echo "image not existing"
docker build . --file docker/Dockerfile --tag smr:$LATEST_SMR_COMMIT --no-cache
docker run -v /var/run/docker.sock:/var/run/docker.sock -v /tmp:/tmp -v /home/qdnqn/testing-smr:/home/smr-agent/.ssh -p 0.0.0.0:1443:1443 --name smr-agent --dns 127.0.0.1 smr:$LATEST_SMR_COMMIT -it