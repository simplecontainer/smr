#!/bin/bash

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
  PRODUCTION=1
  AGENT=""
  DOMAIN=""
  IP=""
  NODE_URL=""
  NODE_PORT="9212"
  CONN_STRING="https://localhost:1443"
  CLIENT_ARGS="--overlayport 0.0.0.0:9212"
  MODE="cluster"
  JOIN=""
  CONTROL_PLANE="0.0.0.0:1443"
  REPOSITORY="simplecontainermanager/smr"
  TAG=$(curl -sL https://raw.githubusercontent.com/simplecontainer/smr/main/version)

  while getopts ":c:e:h:m:a:d:i:j:n:r:p:x:t:z:" option; do
     case $option in
        h) # Display help
           HelpStart && exit;;
        c) # Control plane client connect string
           CONN_STRING=$OPTARG;;
        e) # Expose control plane
           CONTROL_PLANE=$OPTARG;;
        m) #Set mode
           MODE=$OPTARG;;
        j) #Set join
           JOIN=$OPTARG;;
        a) # Set agent
           AGENT=$OPTARG;;
        d) # Set domain
           DOMAIN=$OPTARG;;
        t) # Set tag
           TAG=$OPTARG;;
        i) # Set ip
           IP=$OPTARG;;
        n) # Set Node URL
           NODE_URL=$OPTARG;;
        p) # Set node port
           NODE_PORT=$OPTARG;;
        r) # Set repository
           REPOSITORY=$OPTARG;;
        x) # Set client additional args
           CLIENT_ARGS=$OPTARG;;
        z) # Production or not
           PRODUCTION=$OPTARG;;
        *) # Invalid option
          echo "Invalid option";;
     esac
  done

  echo "Agent name: $AGENT"
  echo "Node URL: $NODE_URL"
  echo "Domain: $DOMAIN"
  echo "Additional arguments: $CLIENT_ARGS"
  echo "Join: $JOIN"

  if [[ ${AGENT} != "" ]]; then
    if [[ ${MODE} == "cluster" ]]; then
      smr node run --image "${REPOSITORY}" --tag "${TAG}" --args="create --agent ${AGENT} --domains ${DOMAIN} --ips ${IP}" --agent "${AGENT}" $CLIENT_ARGS --wait
      smr node run --image "${REPOSITORY}" --tag "${TAG}" --args="start" $CLIENT_ARGS --agent "${AGENT}"

      if [[ $PRODUCTION == "0" ]]; then
        NODE_URL="https://$(docker inspect -f '{{.NetworkSettings.Networks.bridge.IPAddress}}' $AGENT):${NODE_PORT}"
      fi

      smr context connect "${CONN_STRING}" "${HOME}/.ssh/simplecontainer/root.pem" --context "${AGENT}" --wait --y

      if [[ ${JOIN} == "" ]]; then
        sudo nohup smr node cluster start --node "${NODE_URL}" 2>&1 | dd of=~/smr/smr/logs/$AGENT-cluster.log &>/dev/null &
      else
        sudo nohup smr node cluster start --node "${NODE_URL}" --join ${JOIN} 2>&1 | dd of=~/smr/smr/logs/$AGENT-cluster-join.log &>/dev/null &
      fi

      echo "The simplecontainer is started in cluster mode."
    else
      smr node run --image "${REPOSITORY}" --tag "${TAG}" --args="create --agent ${AGENT} --domain ${DOMAIN} --ip ${IP}" --agent "${AGENT}" --wait
      smr node run --image "${REPOSITORY}" --tag "${TAG}" --args="start" $CLIENT_ARGS "${CONTROL_PLANE}" --agent "${AGENT}"

      echo "The simplecontainer is started in single mode."
    fi
  else
    HelpStart
  fi
}

Wait(){
  AGENT=""

  while getopts ":e:h:m:a:d:i:j:n:r:p:x:t:" option; do
     case $option in
        a) # Set agent
           PRODUCTION=$OPTARG;;
        *) # Invalid option
          echo "Invalid option";;
     esac
  done

  smr node wait --agent "${AGENT}"
}

Export(){
  smr context export <<< $1
}

Import(){
  KEY=""

  while read line
  do
    KEY=$line
  done < /dev/stdin

  smr context import "${1}" <<< "${KEY}"
}

Download(){
  which curl &> /dev/null || echo "Please install curl before proceeding with installing smr!" | exit 1
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
      shift
      Download "$@";;
    "start")
      shift
      Start "$@";;
    "wait")
     shift
     Wait "$@";;
    "import")
     shift
     Import "$@";;
    "export")
     shift
     Export "$@";;
    *)
      echo "Unknown command: $COMMAND" ;;
esac
