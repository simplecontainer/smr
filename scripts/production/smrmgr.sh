#!/bin/bash

HelpStart(){
  echo """Usage:

 Eg:
 ./start.sh -a smr-node-1 -d example.com -i 1 -n https://node1.example.com -o 0.0.0.0:9212 -c https://node1.example.com:9212,https:node2.example.com:9212

 Options:
 -a: Node domain
 -d: Domain of agent
"""
}

extract_flag_value() {
  local input="$1"
  local flag="$2"

  # Extract value for the given flag using regex
  if [[ "$input" =~ ($flag)[=[:space:]]([^[:space:]]+) ]]; then
    echo "${BASH_REMATCH[2]}"
  else
    echo ""
  fi
}

#
#
# smrmgr

Manager(){
  NODE=""
  DOMAIN=""
  IP=""
  NODE_DOMAIN=""
  NODE_PORT="1443"
  RAFT_PORT="9212"
  CONN_STRING="https://localhost:1443"
  NODE_ARGS="--listen 0.0.0.0:1443"
  CLIENT_ARGS="--port.control 0.0.0.0:1443 --port.overlay 0.0.0.0:9212"
  MODE="cluster"
  JOIN=false
  PEER=""
  REPOSITORY="quay.io/simplecontainer/smr"
  TAG=$(curl -sL https://raw.githubusercontent.com/simplecontainer/smr/main/version)
  ALLYES=false

  echo "All arguments: $*"

  while getopts ":a:c:d:e:h:i:m:n:p:r:t:x:z:sujy" option; do
    case $option in
      a) # Set agent
         NODE=$OPTARG; ;;
      c) # Control plane client connect string
         CONN_STRING=$OPTARG; ;;
      d) # Set domain
         DOMAIN=$OPTARG; ;;
      e) # Expose control plane
         NODE_ARGS=$OPTARG; ;;
      h) # Display help
         HelpStart && exit; ;;
      i) # Set ip
         IP=$OPTARG; ;;
      j) #Set join
         JOIN="true"; ;;
      m) #Set mode
         MODE=$OPTARG; ;;
      n) # Set Node URL
         NODE_DOMAIN=$OPTARG; ;;
      p) # Set node port
         NODE_PORT=$OPTARG; ;;
      r) # Set repository
         REPOSITORY=$OPTARG; ;;
      t) # Set tag
         TAG=$OPTARG; ;;
      x) # Set client additional args
         CLIENT_ARGS=$OPTARG; ;;
      y) #SSet all yes answer
         ALLYES=true; ;;
      z) PEER=$OPTARG; ;;
      *) # Invalid option
        echo "Invalid option"; ;;
   esac
  done

  if [[ $DOMAIN != "" ]]; then
    if [[ $NODE_DOMAIN == "" ]]; then
      NODE_DOMAIN=$DOMAIN
    fi
  fi

  if [[ $DOMAIN == "" && $NODE_DOMAIN == "" ]]; then
    DOMAIN="localhost"
    NODE_DOMAIN="localhost"
  fi

  NODE_ARGS="--image ${REPOSITORY} --tag ${TAG} --node ${NODE} ${NODE_ARGS}"

  [[ -n "$DOMAIN" ]] && NODE_ARGS+=" --domains ${DOMAIN}"
  [[ -n "$IP" ]] && NODE_ARGS+=" --ips ${IP}"
  [[ -n "$PEER" && "$JOIN" == true ]] && NODE_ARGS+=" --join --peer ${PEER}"

  echo "..Node info....................................................................................................."
  echo "....Agent name:           $NODE"
  echo "....Node:                 $NODE_DOMAIN"
  echo "....Repository:           $REPOSITORY"
  echo "....Tag:                  $TAG"
  echo "....Node args:            $NODE_ARGS"
  echo "....Client args:          $CLIENT_ARGS"

  if [[ $JOIN == "true" ]]; then
    echo "....Join:                 $JOIN"
    echo "....Peer:                 $PEER"
  else
    echo "....Join:                 false"
  fi

  #echo "....smr version:          $(smr version)"
  echo "....ctl version:          $(smrctl version)"
  #echo "....node logs path:       ~/smr/logs/flannel-${NODE}.log"
  echo "................................................................................................................"

  if ! dpkg -s curl &>/dev/null; then
    echo 'please install curl manually'
    exit 1
  fi

  if [[ ${NODE} != "" ]]; then
    if [[ ! $(smr node create --node "${NODE}" --image "${REPOSITORY}" --tag "${TAG}" $NODE_ARGS $CLIENT_ARGS) ]]; then
      echo "failed to create node configuration"
      exit 2
    fi

    touch ~/nodes/${NODE}/logs/cluster.log || (echo "failed to create log file: ~/smr/logs/cluster.log" && exit 2)
    touch ~/nodes/${NODE}/logs/control.log || (echo "failed to create log file: ~/smr/logs/control.log" && exit 2)

    smr node start --node "${NODE}"

    if [[ $NODE_DOMAIN == "localhost" ]]; then
      RAFT_URL="https://$(docker inspect -f '{{.NetworkSettings.Networks.bridge.IPAddress}}' $NODE):${RAFT_PORT}"
    else
      RAFT_URL="https://${NODE_DOMAIN}:${RAFT_PORT}"
    fi

    sudo nohup smr agent start --node "${NODE}" --raft "${RAFT_URL}" --y </dev/null 2>&1 | stdbuf -o0 grep "" > ~/nodes/${NODE}/logs/cluster.log &
    sudo nohup smr agent control --node "${NODE}" </dev/null 2>&1 | stdbuf -o0 grep "" > ~/nodes/${NODE}/logs/control.log &

    echo "tail flannel logs at: tail -f ~/nodes/${NODE}/logs/cluster.log"
    echo "tail control logs at: tail -f ~/nodes/${NODE}/logs/control.log"
    echo "waiting for cluster to be ready..."

    while [ ! -f "$HOME/nodes/${NODE}/logs/control.log" ]; do sleep 0.1; done
    (tail -F "$HOME/nodes/${NODE}/logs/control.log" | grep --line-buffered "cluster started with success" | { read line; echo "cluster started with success"; killall tail; })
  else
    HelpStart
  fi
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
  smr context fetch
}

Download(){
  which curl &> /dev/null || echo "Please install curl before proceeding with installing smr!" | exit 1
  echo "Downloading smr binary and installing it to the /usr/local/bin/smr"
  ARCH=$(uname -p)

  if [[ $ARCH == "x86_64" ]]; then
    ARCH="amd64"
  fi

  VERSION=${2:-$(curl -sL https://raw.githubusercontent.com/simplecontainer/client/main/version)}
  PLATFORM="linux-${ARCH}"

  curl -Lo client https://github.com/simplecontainer/client/releases/download/$VERSION/client-$PLATFORM
  chmod +x client

  sudo mv client /usr/local/bin/smr
}

COMMAND=${1}
shift

case "$COMMAND" in
    "install")
      Download "$@";;
    "start")
      Manager "$@";;
    "wait")
     Wait "$@";;
    "import")
     Import "$@";;
    "export")
     Export "$@";;
    *)
      echo "Unknown command: $COMMAND" ;;
esac