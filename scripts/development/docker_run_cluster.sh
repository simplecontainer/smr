#!/bin/bash
cd "$(dirname "$0")"
cd ../../

echo "Doing work in directory $PWD"

BRANCH="$(git rev-parse --abbrev-ref HEAD)"
LATEST_SMR_COMMIT="$(git rev-parse --short "$BRANCH")"

docker stop smr-agent-1
smr node run --image smr --tag $LATEST_SMR_COMMIT --args="create smr --agent smr-agent-1" --ips 127.0.0.1 --domains localhost,smr-agent-1 --agent smr-agent-1 --wait
smr node run --image smr --tag $LATEST_SMR_COMMIT --args="start" --agent smr-agent-1
smr context connect https://localhost:1443 $HOME/.ssh/simplecontainer/root.pem --context smr-agent-1
smr node cluster start --node 1 --url https://$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' smr-agent-1):9212  --cluster https://$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' smr-agent-1):9212

#smr node run --image smr --tag $LATEST_SMR_COMMIT --args="create smr --agent smr-agent-2" --ips 127.0.0.1  --domains localhost,smr-agent-2 --agent smr-agent-2 --hostport 1444
#smr node run --image smr --tag $LATEST_SMR_COMMIT --args="start" --agent smr-agent-2 --hostport 1444
#smr node cluster add --node 2 --url https://$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' smr-agent-1):9212
#smr context connect https://localhost:1444 $HOME/.ssh/simplecontainer/root.pem --context smr-agent-2 -y
#smr node cluster start --node 2 --url https://$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' smr-agent-2):9212