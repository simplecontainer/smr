#!/bin/bash
cd "$(dirname "$0")"
cd ../../

echo "Doing work in directory $PWD"

BASE_DIR="$PWD"
BRANCH="$(git rev-parse --abbrev-ref HEAD)"
LATEST_SMR_COMMIT="$(git rev-parse --short $BRANCH)"

cd "$BASE_DIR"

echo "***********************************************"
echo "$BASE_DIR/../implementations/$DIRNAME"
echo "***********************************************"

go build -ldflags '-s -w' || exit 1

for dir in implementations/*/
do
    DIR=${dir%*/}
    DIRNAME="${DIR##*/}"

    echo "***********************************************"
    echo "$BASE_DIR/../implementations/$DIRNAME"
    echo "***********************************************"

    cd "$BASE_DIR/implementations/$DIRNAME"
    rm -rf *.so

    CGO_ENABLED=1 go build -ldflags '-s -w' --buildmode=plugin || exit 1
done

cd "$BASE_DIR"