#!/bin/bash
cd "$(dirname "$0")"

set -e

echo "Doing work in directory $PWD"

sudo pkill -f smr

BRANCH="$(git rev-parse --abbrev-ref HEAD)"
TAG="$(git rev-parse --short "$BRANCH")"

docker stop smr-development-node1-1 smr-development-node-2 smr-development-node-3 || echo
docker rm smr-development-node-1 smr-development-node-2 smr-development-node-3 || echo

rm -rf ~/smr/.ssh || echo

rm -rf ~/nodes/smr-development-node-1  || echo
rm -rf ~/nodes/smr-development-node-2  || echo
rm -rf ~/nodes/smr-development-node-3  || echo

../production/smrmgr.sh start -n node-one -d localhost -i smr -t "$TAG" -c "--port.control 0.0.0.0:1443 --port.etcd 2379 --port.overlay 0.0.0.0:9212" --
CLUSTER_DOMAIN_1="https://$(smr node ip --node node-one):1443"

JOIN_ARGS=$(smr agent export --api $CLUSTER_DOMAIN_1 --node node-one)
smr agent import --node node-two -y $JOIN_ARGS

../production/smrmgr.sh start -n node-two -d localhost -i smr -t "$TAG" -j -c "--port.control 0.0.0.0:1444 --port.etcd 2380 --port.overlay 0.0.0.0:9213" --
CLUSTER_DOMAIN_2="$(docker inspect -f '{{.NetworkSettings.Networks.bridge.IPAddress}}' node-two):1444"

#../production/smrmgr.sh start -n smr-development-node-3 -d localhost -c '--port.control 0.0.0.0:1445 --port.etcd 2381 --port.overlay 0.0.0.0:9214' -r smr -t $TAG -j -p $CLUSTER_DOMAIN_1
#CLUSTER_DOMAIN_2="$(docker inspect -f '{{.NetworkSettings.Networks.bridge.IPAddress}}' smr-development-node-2):1444"
