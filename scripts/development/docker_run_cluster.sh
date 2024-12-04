#!/bin/bash
cd "$(dirname "$0")"
cd ../../

echo "Doing work in directory $PWD"

BRANCH="$(git rev-parse --abbrev-ref HEAD)"
LATEST_SMR_COMMIT="$(git rev-parse --short "$BRANCH")"

docker stop smr-agent-1 smr-agent-2
docker rm smr-agent-1 smr-agent-2

rm -rf ~/.smr-agent-1
rm -rf ~/.smr-agent-2

smr node run --image smr --tag $LATEST_SMR_COMMIT --args="create --agent smr-agent-1" --agent smr-agent-1 --wait
smr node run --image smr --tag $LATEST_SMR_COMMIT --args="start" --agent smr-agent-1 --overlayport 0.0.0.0:9212
smr context connect https://localhost:1443 $HOME/.ssh/simplecontainer/root.pem --context smr-agent-1 -y --wait

CLUSTER_DOMAIN_1="https://$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' smr-agent-1):9212"
smr node cluster start --node 1 --url $CLUSTER_DOMAIN_1  --cluster $CLUSTER_DOMAIN_1

smr node run --image smr --tag $LATEST_SMR_COMMIT --args="create --agent smr-agent-2" --hostport 1444 --etcdport 2380 --agent smr-agent-2 --wait
smr node run --image smr --tag $LATEST_SMR_COMMIT --args="start" --hostport 1444 --etcdport 2380 --overlayport 0.0.0.0:9213 --agent smr-agent-2

CLUSTER_DOMAIN_2="https://$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' smr-agent-2):9212"
smr node cluster add --node 2 --url $CLUSTER_DOMAIN_2

smr context connect https://localhost:1444 $HOME/.ssh/simplecontainer/root.pem --context smr-agent-2 -y --wait
smr node cluster start --node 2 --url $CLUSTER_DOMAIN_2 --cluster $CLUSTER_DOMAIN_1,$CLUSTER_DOMAIN_2 --join