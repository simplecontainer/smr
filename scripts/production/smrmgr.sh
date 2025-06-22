#!/bin/bash

HelpStart(){
  echo """
Usage: smrmgr.sh start [options]

Options:
  -n <node>         Set node name (e.g., node-1)
  -d <domain>       Set domain (e.g., example.com)
  -a <ip address>   Set IP address (e.g., 192.168.1.10)
  -c <args>         Set additional client arguments (e.g., "--foo bar")
  -i <image>        Set Docker image (e.g., myrepo/myimage)
  -t <tag>          Set Docker image tag (e.g., latest)
  -j                Join an existing cluster (no value needed)
  -p <peer>         Set peer address (e.g., 192.168.1.20)
  -s                Install as a systemd service (no value needed)
  -h                Show this help message and exit

Examples:
  smrmgr.sh -n node-1 -d mydomain.com -a 10.0.0.1 -i myrepo/myimage -t latest -s
  smrmgr.sh -n node-2 -j -p 10.0.0.1
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
  TAG=$(curl -sL https://raw.githubusercontent.com/simplecontainer/smr/refs/heads/main/cmd/smr/version)
  SERVICE="false"

  echo "All arguments: $*"

  while getopts ":a:c:d:h:i:n:p:t:j:s" option; do
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
      s) #Install as service
         SERVICE="true" ;;
      h) # Display help
         HelpStart && exit; ;;
      *) # Invalid option
        echo "Invalid option"; ;;
   esac
  done

  if [[ ${NODE} == "" ]]; then
    NODE="simplecontainer-node-1"
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
  echo "....Service install:      $SERVICE"

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

  if ! dpkg -s docker-ce &>/dev/null; then
    echo 'please install docker manually'
    exit 1
  fi

  if [[ ${NODE} != "" ]]; then
    if [[ ! $(smr node create --node "${NODE}" $NODE_ARGS $CLIENT_ARGS) ]]; then
      echo "failed to create node configuration"
      exit 2
    fi

    touch ~/nodes/${NODE}/logs/cluster.log || (echo "failed to create log file: ~/smr/logs/cluster.log" && exit 2)
    touch ~/nodes/${NODE}/logs/control.log || (echo "failed to create log file: ~/smr/logs/control.log" && exit 2)

    if [[ $DOMAIN == "localhost" ]]; then
      RAFT_URL="https://$(docker inspect -f '{{.NetworkSettings.Networks.bridge.IPAddress}}' $NODE):9212"
    else
      RAFT_URL="https://${DOMAIN}:9212"
    fi

    if [[ $SERVICE == "false" ]]; then
      Start "${NODE}" "${RAFT_URL}"

      echo "tail flannel logs at: tail -f ~/nodes/${NODE}/logs/cluster.log"
      echo "tail control logs at: tail -f ~/nodes/${NODE}/logs/control.log"
      echo "waiting for cluster to be ready..."

      smr agent events --wait cluster_started --node "$NODE"
    fi
  else
    HelpStart
  fi
}

Start(){
  smr node start --node "${1}" -y

  sudo nohup smr agent start --node "${1}" --raft "${2}" </dev/null 2>&1 | stdbuf -o0 grep "" > ~/nodes/${1}/logs/cluster.log &
  sudo nohup smr agent control --node "${1}" </dev/null 2>&1 | stdbuf -o0 grep "" > ~/nodes/${1}/logs/control.log &
}

Stop(){
  smr agent drain
  smr events --wait drain_success
  smr node clean --node ${1}

  sudo smr agent stop
}

Download(){
  curl --version 2&>1 /dev/null || echo "Please install curl before proceeding with installing smr!" | exit 1

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
    "stop")
      Stop "$@";;
    "service-start")
      Start "$@";;
    "service-stop")
      Stop "$@";;
    *)
      echo "Available commands are: install, start, stop, service-start, service-stop" ;;
esac