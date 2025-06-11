#!/bin/bash

HelpStart(){
  echo """Usage:

 Eg:
 ./start.sh -n smr-node-1 -d example.com

 Options:
 -n: Node name
 -d: Node domain
 -a: Node IP address
 -r Raft port
"""
}

#
#
# smrmgr

Manager(){
  NODE=""
  DOMAIN=""
  IP=""
  NODE_ARGS="--listen 0.0.0.0:1443"
  CLIENT_ARGS="--port.control 0.0.0.0:1443 --port.overlay 0.0.0.0:9212"
  JOIN=false
  PEER=""
  IMAGE="quay.io/simplecontainer/smr"
  TAG=$(curl -sL https://raw.githubusercontent.com/simplecontainer/smr/main/version)

  echo "All arguments: $*"

  while getopts ":a:c:d:h:i:n:p:t:j" option; do
    case $option in
      n) # Set node
         NODE=$OPTARG; ;;
      d) # Set domain
         DOMAIN=$OPTARG; ;;
      a) # Set ip addr
         IP=$OPTARG; ;;
      c) # Set client args
         CLIENT_ARGS=$OPTARG; ;;
      i) # Set repository/image
         IMAGE=$OPTARG; ;;
      t) # Set tag
         TAG=$OPTARG; ;;
      j) #Set join
         JOIN="true"; ;;
      p) #Set peer
         PEER=$OPTARG; ;;
      h) # Display help
         HelpStart && exit; ;;
      *) # Invalid option
        echo "Invalid option"; ;;
   esac
  done

  if [[ ${NODE} == "" ]]; then
    NODE="simplecontainer-node"
  fi

  if [[ $DOMAIN == "" ]]; then
    DOMAIN="localhost"
  fi

  NODE_ARGS="--image ${IMAGE} --tag ${TAG} --node ${NODE} ${NODE_ARGS}"

  [[ -n "$DOMAIN" ]] && NODE_ARGS+=" --domain ${DOMAIN}"
  [[ -n "$IP" ]] && NODE_ARGS+=" --ip ${IP}"
  [[ -n "$PEER" && "$JOIN" == true ]] && NODE_ARGS+=" --join --peer ${PEER}"

  echo "..Node info....................................................................................................."
  echo "....Agent name:           $NODE"
  echo "....Node:                 $DOMAIN"
  echo "....Image:                $IMAGE"
  echo "....Tag:                  $TAG"
  echo "....Node args:            $NODE_ARGS"
  echo "....Client args:          $CLIENT_ARGS"

  if [[ $JOIN == "true" ]]; then
    echo "....Join:                 $JOIN"
    echo "....Peer:                 $PEER"
  else
    echo "....Join:                 false"
  fi

  echo "....smr version:          $(smr version)"
  echo "....ctl version:          $(smrctl version)"
  echo "................................................................................................................"

  if ! dpkg -s curl &>/dev/null; then
    echo 'please install curl manually'
    exit 1
  fi

  if [[ ${NODE} != "" ]]; then
    if [[ ! $(smr node create --node "${NODE}" $NODE_ARGS $CLIENT_ARGS) ]]; then
      echo "failed to create node configuration"
      exit 2
    fi

    touch ~/nodes/${NODE}/logs/cluster.log || (echo "failed to create log file: ~/smr/logs/cluster.log" && exit 2)
    touch ~/nodes/${NODE}/logs/control.log || (echo "failed to create log file: ~/smr/logs/control.log" && exit 2)

    smr node start --node "${NODE}" -y

    if [[ $DOMAIN == "localhost" ]]; then
      RAFT_URL="https://$(docker inspect -f '{{.NetworkSettings.Networks.bridge.IPAddress}}' $NODE):9212"
    else
      RAFT_URL="https://${DOMAIN}:9212"
    fi

    sudo nohup smr agent start --node "${NODE}" --raft "${RAFT_URL}" </dev/null 2>&1 | stdbuf -o0 grep "" > ~/nodes/${NODE}/logs/cluster.log &
    sudo nohup smr agent control --node "${NODE}" </dev/null 2>&1 | stdbuf -o0 grep "" > ~/nodes/${NODE}/logs/control.log &

    echo "tail flannel logs at: tail -f ~/nodes/${NODE}/logs/cluster.log"
    echo "tail control logs at: tail -f ~/nodes/${NODE}/logs/control.log"
    echo "waiting for cluster to be ready..."

    smr agent events --wait cluster_started --node "$NODE"
  else
    HelpStart
  fi
}

Download(){
  which curl &> /dev/null || { echo "Please install curl before proceeding with installing smr!"; exit 1; }

  ARCH=$(detect_arch)
  PLATFORM="linux-${ARCH}"

  VERSION_SMR=${2:-$(curl -sL https://raw.githubusercontent.com/simplecontainer/smr/refs/heads/main/cmd/smr/version --fail)}
  VERSION_CTL=${2:-$(curl -sL https://raw.githubusercontent.com/simplecontainer/smr/refs/heads/main/cmd/smrctl/version)}

  echo "Downloading smr:$VERSION_SMR and smrctl:$VERSION_CTL binary. They will be installed at /usr/local/bin/smr"
  echo "Downloading: https://github.com/simplecontainer/smr/releases/download/smr-$VERSION_SMR/smr-$PLATFORM"

  curl -Lo smr https://github.com/simplecontainer/smr/releases/download/smr-$VERSION_SMR/smr-$PLATFORM --fail
  chmod +x smr

  if ! ./smr --help > /dev/null 2>&1 && ! ./smr --version > /dev/null 2>&1; then
    echo "smr is not executable or cannot run."
    exit 1
  fi

  echo "Downloading https://github.com/simplecontainer/smr/releases/download/smrctl-$VERSION_CTL/smrctl-$PLATFORM"
  curl -Lo smrctl https://github.com/simplecontainer/smr/releases/download/smrctl-$VERSION_CTL/smrctl-$PLATFORM --fail
  chmod +x smrctl

  if ! ./smrctl --help > /dev/null 2>&1 && ! ./smrctl --version > /dev/null 2>&1; then
    echo "smrctl is not executable or cannot run."
    exit 1
  fi

  sudo mv smr /usr/local/bin/smr
  sudo mv smrctl /usr/local/bin/smrctl

  echo "smr and smrctl have been successfully installed to /usr/local/bin."
}


detect_arch() {
  ARCH=""

  if command -v uname >/dev/null 2>&1; then
    ARCH="$(uname -m 2>/dev/null)"
  fi

  if [ -z "$ARCH" ] && command -v dpkg >/dev/null 2>&1; then
    ARCH="$(dpkg --print-architecture 2>/dev/null)"
  fi

  if [ -z "$ARCH" ] && command -v arch >/dev/null 2>&1; then
    ARCH="$(arch 2>/dev/null)"
  fi

  case "$ARCH" in
    x86_64 | amd64)
      echo "amd64"
      ;;
    aarch64 | arm64)
      echo "arm64"
      ;;
    armv8* | armv7* | armv6* | armhf)
      echo "arm64"  # adjust here if you want separate armv7l, etc.
      ;;
    *)
      echo "Error: Unknown or unsupported architecture: '$ARCH'" >&2
      exit 1
      ;;
  esac
}

COMMAND=${1}
shift

case "$COMMAND" in
    "install")
      Download "$@";;
    "start")
      Manager "$@";;
    *)
      echo "Available commands are: install and start" ;;
esac