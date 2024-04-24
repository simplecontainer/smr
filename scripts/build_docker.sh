#!/bin/bash
cd "$(dirname "$0")"

echo "Starting in directory $PWD"

BASE_DIR="../$PWD"

for dir in implementations/*/
do
    DIR=${dir%*/}
    DIRNAME="${DIR##*/}"

    cd "$BASE_DIR/implementations/$DIRNAME"
    go build --buildmode=plugin
done

cd "$BASE_DIR"

for dir in operators/*/
do
    DIR=${dir%*/}
    DIRNAME="${DIR##*/}"

    cd "$BASE_DIR/operators/$DIRNAME"
    go build --buildmode=plugin
done

cd "$BASE_DIR"

go build

docker stop $(docker ps -q)
docker rm $(docker ps -aq)
docker build . --file docker/Dockerfile --tag smr:0.0.1
docker run -v /var/run/docker.sock:/var/run/docker.sock -v /tmp:/tmp -p 127.0.0.1:8080:8080 --name smr-agent -it smr:0.0.1