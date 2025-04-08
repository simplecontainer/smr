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

Manager(){
  NODE=""
  DOMAIN=""
  IP=""
  NODE_DOMAIN=""
  NODE_PORT="9212"
  CONN_STRING="https://localhost:1443"
  CLIENT_ARGS="--static.overlayport 0.0.0.0:9212"
  MODE="cluster"
  JOIN=false
  PEER=""
  CONTROL_PLANE="0.0.0.0:1443"
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
         CONTROL_PLANE=$OPTARG; ;;
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

  echo "..Node info....................................................................................................."
  echo "....Agent name:           $NODE"
  echo "....Node:                 $NODE_DOMAIN"
  echo "....Additional args:      $CLIENT_ARGS"
  echo "....Restart:              $RESTART"
  echo "....Upgrade:              $UPGRADE"

  if [[ $JOIN == "true" ]]; then
  echo "....Join:                 true"
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
      if [[ ${JOIN} == true ]]; then
        ID=$(smr node create --node "${NODE}" --static.image "${REPOSITORY}" --static.tag "${TAG}" $CLIENT_ARGS --args="create --image ${REPOSITORY} --tag ${TAG} --join --peer ${PEER} --node ${NODE}  --port ${CONTROL_PLANE} --domains ${DOMAIN} --ips ${IP}" --w exited)
      else
        ID=$(smr node create --node "${NODE}" --static.image "${REPOSITORY}" --static.tag "${TAG}" $CLIENT_ARGS --args="create --image ${REPOSITORY} --tag ${TAG} --node ${NODE}  --port ${CONTROL_PLANE} --domains ${DOMAIN} --ips ${IP}" --w exited)
      fi

      EXIT_CODE=${?}

      if [[ ${EXIT_CODE} != 0 ]]; then
        echo $ID
        exit ${EXIT_CODE}
      fi

      echo "Configuration created with success - configuration container id: $ID"

      smr node rename --node "${NODE}" "${NODE}-create-${ID}" || exit 3
      smr node run --node "${NODE}" --args="start" --w running

      if [[ $NODE_DOMAIN == "localhost" ]]; then
        NODE_DOMAIN="https://$(docker inspect -f '{{.NetworkSettings.Networks.bridge.IPAddress}}' $NODE):${NODE_PORT}"
      else
        NODE_DOMAIN="https://${NODE_DOMAIN}:${NODE_PORT}"
      fi

      while :
      do
        if smr context connect "${CONN_STRING}" "${HOME}/.ssh/simplecontainer/${NODE}.pem" --context "${NODE}" --y; then
          break
        else
          echo "Failed to connect to siplecontainer, trying again in 1 second"
          sleep 1
        fi
      done

      sudo nohup smr node cluster join --node "$NODE" --api "${NODE_DOMAIN}" </dev/null 2>&1 | stdbuf -o0 grep "" > ~/smr/logs/flannel-${NODE}.log &

      echo "The simplecontainer is started in cluster mode."
    else
      smr node run --node "${NODE}" "${CONTROL_PLANE}" $CLIENT_ARGS --image "${REPOSITORY}" --tag "${TAG}" --image "${REPOSITORY}" --tag "${TAG}" --args="create --node ${NODE} --domain ${DOMAIN} --ip ${IP}" --wait exited
      smr node run --node "$NODE" --args="start"

      sudo nohup smr node cluster join --node "$NODE" --api "${NODE_DOMAIN}" </dev/null 2>&1 | stdbuf -o0 grep "" > ~/smr/logs/flannel-${NODE}.log &

      echo "The simplecontainer is started in single mode."
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
