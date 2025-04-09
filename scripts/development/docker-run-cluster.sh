#!/bin/bash
cd "$(dirname "$0")"

set -e

echo "Doing work in directory $PWD"

sudo pkill -f smr

BRANCH="$(git rev-parse --abbrev-ref HEAD)"
TAG="$(git rev-parse --short "$BRANCH")"

docker stop smr-agent-1 smr-agent-2 smr-agent-3 || echo
docker rm smr-agent-1 smr-agent-2 smr-agent-3 || echo

rm -rf ~/.smr-agent-1  || echo
rm -rf ~/.smr-agent-2  || echo
rm -rf ~/.smr-agent-3  || echo

rm -rf ~/smr/config/smr-agent-1.yaml || echo
rm -rf ~/smr/config/smr-agent-2.yaml || echo
rm -rf ~/smr/config/smr-agent-3.yaml || echo

../production/smrmgr.sh start -a smr-agent-1 -d localhost -c https://localhost:1443 -p 9212 -m cluster -x '--static.overlayport 0.0.0.0:9212' -r smr -t $TAG
CLUSTER_DOMAIN_1="https://$(docker inspect -f '{{.NetworkSettings.Networks.bridge.IPAddress}}' smr-agent-1):1443"

sleep 5

../production/smrmgr.sh start -a smr-agent-2 -d localhost -c https://localhost:1444 -p 9213 -m cluster -x '--static.hostport 1444 --static.etcdport 2380 --static.overlayport 0.0.0.0:9213' -r smr -t $TAG -j -z $CLUSTER_DOMAIN_1
CLUSTER_DOMAIN_2="$(docker inspect -f '{{.NetworkSettings.Networks.bridge.IPAddress}}' smr-agent-2):1444"

#sleep 5
#
#../production/smrmgr.sh start -a smr-agent-3 -d localhost -c https://localhost:1445 -p 9214 -m cluster -j -x '--static.hostport 1445 --static.etcdport 2381 --static.overlayport 0.0.0.0:9214' -r smr -t $TAG