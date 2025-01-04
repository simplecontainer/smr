#!/bin/bash
cd "$(dirname "$0")"

echo "Doing work in directory $PWD"

BRANCH="$(git rev-parse --abbrev-ref HEAD)"
TAG="$(git rev-parse --short "$BRANCH")"

AGENT_NAME="smr-agent-1"

docker stop $AGENT_NAME
docker rm $AGENT_NAME

rm -rf ~/.$AGENT_NAME

../production/smrmgr.sh start -a $AGENT_NAME -r smr -t $TAG