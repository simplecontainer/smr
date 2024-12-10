#!/bin/bash

export AGENT_DOMAIN="https://localhost:1443"
export REPOSITORY="simeplcontainermanager/smr"
export TAG=$(curl -s https://raw.githubusercontent.com/simplecontainer/smr/main/version)

HelpStart(){
  echo """Usage:

 Eg:
 ./start.sh -a https://localhost:1443 -d example.com -i 1 -n https://node1.example.com -o 0.0.0.0:9212 -c https://node1.example.com:9212,https:node2.example.com:9212

 Options:
 -a: Agent domain
 -m: Mode: single or cluster
 -d: Domain where agent is exposed used for certificates
 -i: Node ID; Should be increased monotonically
 -n: Node URL
 -c: Cluster URLs
 -o: Overlay port default is 0.0.0.0:9212
"""
}

Start(){
  AGENT=""
  DOMAIN=""
  IP=""
  NODE_ID=""
  NODE_URL=""
  CLUSTER=""
  OVERLAY="0.0.0.0:9212"
  MODE="cluster"
  JOIN=""
  CONTROL_PLANE="0.0.0.0:1443"

  while getopts ":ehmadijnco:" option; do
     case $option in
        e) #Expose control plane
           CONTROL_PLANE=$OPTARG;;
        h) #Display help
           HelpStart && exit;;
        m) #Set mode
           MODE=$OPTARG;;
        j) #Set join
           JOIN=$OPTARG;;
        a) # Set agent
           AGENT=$OPTARG;;
        d) # Set domain
           DOMAIN=$OPTARG;;
        p) # Set ip
           IP=$OPTARG;;
        i) # Set Node ID
           NODE_ID=$OPTARG;;
        n) # Set Node URL
           NODE_URL=$OPTARG;;
        c) # Set cluster
           CLUSTER=$OPTARG;;
        o) # Set overlay
           OVERLAY=$OPTARG;;
        *) # Default -> Show help
           HelpStart && exit;;
     esac
  done

  if [[ ${AGENT} != "" ]]; then
    if [[ ${MODE} == "cluster" ]]; then
      smr node run --image "${REPOSITORY}" --tag "${TAG}" --args="create --agent ${AGENT} --domains ${DOMAIN}" --ip "${IP}" --agent "${AGENT}" --wait
      smr node run --image "${REPOSITORY}" --tag "${TAG}" --args="start" --overlayport "${OVERLAY}" --agent "${AGENT}"

      smr context connect $AGENT_DOMAIN "${HOME}/.ssh/simplecontainer/root.pem" --context "${AGENT}" --y

      if [[ ${JOIN} == "" ]]; then
        sudo nohup smr node cluster start --node "${NODE_ID}" --url "${NODE_URL}" --cluster "${CLUSTER}"
      else
        smr context switch "${JOIN}" --y
        smr smr node cluster add --node "${NODE_ID}" --url "${NODE_URL}"

        smr context switch "${AGENT}" --y
        sudo nohup smr node cluster start --node "${NODE_ID}" --url "${NODE_URL}" --cluster "${CLUSTER}" --join
      fi

      echo "The simplecontainer is started in cluster mode."
    else
      smr node run --image "${REPOSITORY}" --tag "${TAG}" --args="create --agent ${AGENT} --domain ${DOMAIN} --ip ${IP}" --agent "${AGENT}" --wait
      smr node run --image "${REPOSITORY}" --tag "${TAG}" --args="start" "${CONTROL_PLANE}" --agent "${AGENT}"

      echo "The simplecontainer is started in single mode."
    fi
  else
    HelpStart
  fi
}

Download(){
  which curl &> /dev/null || echo "Please install curl before proceeding with installing smr!" && exit
  echo "Downloading smr binary and installing it to the /usr/local/bin/smr"
  ARCH=$(uname -p)

  if [[ $ARCH == "x86_64" ]]; then
    ARCH="amd64"
  fi

  LATEST_VERSION=$(curl -sL https://raw.githubusercontent.com/simplecontainer/client/main/version)
  PLATFORM="linux-${ARCH}"

  curl -Lo client https://github.com/simplecontainer/client/releases/download/$LATEST_VERSION/client-$PLATFORM
  chmod +x client

  sudo mv client /usr/local/bin/smr
}

COMMAND=${1}

case "$COMMAND" in
    "install")
        Download;;

    "start")
      Start;;

    *)
        echo "Unknown command: $command" ;;
esac
