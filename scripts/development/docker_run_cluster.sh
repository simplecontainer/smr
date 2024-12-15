#!/bin/bash
cd "$(dirname "$0")"

echo "Doing work in directory $PWD"

BRANCH="$(git rev-parse --abbrev-ref HEAD)"
TAG="$(git rev-parse --short "$BRANCH")"

docker stop smr-agent-1 smr-agent-2
docker rm smr-agent-1 smr-agent-2

rm -rf ~/.smr-agent-1
rm -rf ~/.smr-agent-2

../production/smrmgr.sh start -a smr-agent-1 -c https://localhost:1443 -p 9212 -m cluster -x '--overlayport 0.0.0.0:9212' -r smr -t $TAG -z 0
CLUSTER_DOMAIN_1="https://$(docker inspect -f '{{.NetworkSettings.Networks.bridge.IPAddress}}' smr-agent-1):1443"

EXPORTED_CONTEXT=$(../production/smrmgr.sh export $CLUSTER_DOMAIN_1)
../production/smrmgr.sh import $EXPORTED_CONTEXT <<< $(cat $HOME/smr/smr/contexts/smr-agent-1.key)

smr context fetch
sleep 5

../production/smrmgr.sh start -a smr-agent-2 -c https://localhost:1444 -p 9213 -m cluster -x '--hostport 1444 --etcdport 2380 --overlayport 0.0.0.0:9213' -r smr -t $TAG -j $CLUSTER_DOMAIN_1 -z 0
