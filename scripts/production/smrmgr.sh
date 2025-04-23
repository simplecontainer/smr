#!/bin/bash

HelpStart(){
  echo """Usage:

 Eg:
 ./start.sh -a https://localhost:1443 -d example.com -i 1 -n https://node1.example.com -o 0.0.0.0:9212 -c https://node1.example.com:9212,https:node2.example.com:9212

 Options:
 -a: Agent domain
 -m: Mode: standalone or cluster
 -d: Domain of agent
 -n: Node URL - if node URL is different than domain of agent
 -c: Cluster URLs
 -o: Overlay port default is 0.0.0.0:9212
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
  NODE_ARGS="--port 0.0.0.0:1443"
  CLIENT_ARGS="--dynamic.hostport 0.0.0.0:1443 --dynamic.overlayport 0.0.0.0:9212"
  MODE="cluster"
  JOIN=false
  PEER=""
  REPOSITORY="quay.io/simplecontainer/smr"
  TAG=$(curl -sL https://raw.githubusercontent.com/simplecontainer/smr/main/version)
  ALLYES=false

  echo "All arguments: $@"

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

  # Domain not known at the top so append it here!

  if [[ ${JOIN} == true ]]; then
    NODE_ARGS="--image ${REPOSITORY} --tag ${TAG} --node ${NODE} ${NODE_ARGS} --domains ${DOMAIN} --ips ${IP} --join --peer ${PEER}"
  else
    NODE_ARGS="--image ${REPOSITORY} --tag ${TAG} --node ${NODE} ${NODE_ARGS} --domains ${DOMAIN} --ips ${IP}"
  fi

  echo "..Node info....................................................................................................."
  echo "....Agent name:           $NODE"
  echo "....Node:                 $NODE_DOMAIN"
  echo "....Repository:           $REPOSITORY"
  echo "....Tag:                  $TAG"
  echo "....Mode:                 $MODE"
  echo "....Additional args:      $CLIENT_ARGS"

  if [[ $JOIN == "true" ]]; then
    echo "....Join:                 $JOIN"
    echo "....Peer:                 $PEER"
  else
    echo "....Join:                 false"
  fi

  echo "....cli version:          $(smr version)"
  echo "....node logs path:       ~/smr/logs/flannel-${NODE}.log"
  echo "................................................................................................................"

  touch ~/smr/logs/flannel-${NODE}.log || (echo "Failed to create log file: ~/smr/logs/flannel-${NODE}.log" && exit 2)

  if ! dpkg -s curl &>/dev/null; then
    echo 'please install curl manually'
    exit 1
  fi

  if [[ ${NODE} != "" ]]; then
    if [[ ${MODE} == "cluster" ]]; then
      ID=$(smr node create \
        --node "${NODE}" \
        --static.image "${REPOSITORY}" \
        --static.tag "${TAG}" \
        $CLIENT_ARGS \
        --args="create ${NODE_ARGS}" \
        --w exited)

      EXIT_CODE=${?}

      if [[ ${EXIT_CODE} != 0 ]]; then
        echo $ID
        exit ${EXIT_CODE}
      fi

      echo "Configuration created with success - configuration container id: $ID"

      smr node rename --node "${NODE}" "${NODE}-create-${ID}" || exit 3
      smr node run --node "${NODE}" --args="start" --w running

      if [[ $NODE_DOMAIN == "localhost" ]]; then
        RAFT_URL="https://$(docker inspect -f '{{.NetworkSettings.Networks.bridge.IPAddress}}' $NODE):${RAFT_PORT}"
      else
        RAFT_URL="https://${NODE_DOMAIN}:${RAFT_PORT}"
      fi

      echo "Attemp to connect to the simplecontainer node and save context."

      while :
      do
        if smr context connect "${CONN_STRING}" "${HOME}/.ssh/simplecontainer/${NODE}.pem" --context "${NODE}" --y; then
          break
        else
          echo "Failed to connect to simplecontainer node, trying again in 1 second..."
          sleep 1
        fi
      done

      sudo nohup smr node cluster join --node "$NODE" --raft $RAFT_URL </dev/null 2>&1 | stdbuf -o0 grep "" > ~/smr/logs/flannel-${NODE}.log &
      echo "The simplecontainer node is started."
    fi
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