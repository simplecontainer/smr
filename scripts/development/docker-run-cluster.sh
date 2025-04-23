#!/bin/bash
cd "$(dirname "$0")"

set -e

echo "Doing work in directory $PWD"

sudo pkill -f smr

BRANCH="$(git rev-parse --abbrev-ref HEAD)"
TAG="$(git rev-parse --short "$BRANCH")"

docker stop smr-development-node1-1 smr-development-node-2 smr-development-node-3 || echo
docker rm smr-development-node-1 smr-development-node-2 smr-development-node-3 || echo

rm -rf ~/.smr-development-node-1  || echo
rm -rf ~/.smr-development-node-2  || echo
rm -rf ~/.smr-development-node-3  || echo

rm -rf ~/smr/config/smr-development-node-1.yaml || echo
rm -rf ~/smr/config/smr-development-node-2.yaml || echo
rm -rf ~/smr/config/smr-development-node-3.yaml || echo

../production/smrmgr.sh start -a smr-development-node-1 -d localhost -c https://localhost:1443 -x '--dynamic.hostport 0.0.0.0:1443 --dynamic.etcdport 2379 --dynamic.overlayport 0.0.0.0:9212' -r smr -t $TAG
CLUSTER_DOMAIN_1="https://$(docker inspect -f '{{.NetworkSettings.Networks.bridge.IPAddress}}' smr-development-node-1):1443"

sleep 5

../production/smrmgr.sh start -a smr-development-node-2 -d localhost -c https://localhost:1444 -x '--dynamic.hostport 0.0.0.0:1444 --dynamic.etcdport 2380 --dynamic.overlayport 0.0.0.0:9213' -r smr -t $TAG -j -z $CLUSTER_DOMAIN_1
CLUSTER_DOMAIN_2="$(docker inspect -f '{{.NetworkSettings.Networks.bridge.IPAddress}}' smr-development-node-2):1444"

#sleep 5
#../production/smrmgr.sh start -a smr-development-node-3 -d localhost -c https://localhost:1445 -p 9214 -m cluster -x '--dynamic.hostport 0.0.0.0:1445 --dynamic.etcdport 2381 --dynamic.overlayport 0.0.0.0:9214' -r smr -t $TAG -j -z $CLUSTER_DOMAIN_1
#CLUSTER_DOMAIN_2="$(docker inspect -f '{{.NetworkSettings.Networks.bridge.IPAddress}}' smr-development-node-2):1444"
