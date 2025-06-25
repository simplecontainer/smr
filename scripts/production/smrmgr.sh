#!/bin/bash
set -e

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

  # With value on the left side with colon!
  # Without value on right side without colon!
  while getopts "a:c:d:h:i:n:p:t:js" option; do
  case $option in
      n) NODE=$OPTARG ;;
      d) DOMAIN=$OPTARG ;;
      a) IP=$OPTARG ;;
      c) CLIENT_ARGS=$OPTARG ;;
      i) IMAGE=$OPTARG ;;
      t) TAG=$OPTARG ;;
      j) JOIN="true" ;;
      p) PEER=$OPTARG ;;
      s) SERVICE="true" ;;
      h) HelpStart && exit ;;
      \?) echo "Invalid option: -$OPTARG" ;;
      :) echo "Option -$OPTARG requires an argument." ;;
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
  echo "....Domain:               $DOMAIN"
  echo "....IP:                   $IP"
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

  curl --version > /dev/null 2>&1 || echo "Please install curl before proceeding with installing smr!" | exit 1
  docker --version > /dev/null  2>&1 || echo "Please install docker-ce before proceeding with installing smr!" | exit 1

  if [[ ${NODE} != "" ]]; then
    if [[ ! $(smr node create --node "${NODE}" $NODE_ARGS $CLIENT_ARGS) ]]; then
      echo "failed to create node configuration"
      exit 2
    fi

    touch ~/nodes/${NODE}/logs/cluster.log || (echo "failed to create log file: ~/smr/logs/cluster.log" && exit 2)
    touch ~/nodes/${NODE}/logs/control.log || (echo "failed to create log file: ~/smr/logs/control.log" && exit 2)

    if [[ $SERVICE == "false" ]]; then
      Start

      echo "tail flannel logs at: tail -f ~/nodes/${NODE}/logs/cluster.log"
      echo "tail control logs at: tail -f ~/nodes/${NODE}/logs/control.log"
      echo "waiting for cluster to be ready..."

      smr agent events --wait cluster_started --node "$NODE"
    fi

    SaveCurrentEnvToFile "$HOME/nodes/.env"
  else
    HelpStart
  fi
}

Start(){
  smr node start --node "${NODE}" -y

  if [[ $DOMAIN == "localhost" ]]; then
    RAFT_URL="https://$(docker inspect -f '{{.NetworkSettings.Networks.bridge.IPAddress}}' $NODE):9212"
  else
    RAFT_URL="https://${DOMAIN}:9212"
  fi

  sudo nohup smr agent start --node "${NODE}" --raft "${RAFT_URL}" </dev/null 2>&1 | stdbuf -o0 grep "" > ~/nodes/${NODE}/logs/cluster.log &
  nohup smr agent control --node "${NODE}" </dev/null 2>&1 | stdbuf -o0 grep "" > ~/nodes/${NODE}/logs/control.log &
}

Stop(){
  smr agent drain --node $NODE
  smr agent events --node ${NODE} --wait drain_success
  smr node clean --node ${NODE}

  sudo smr agent stop agent
  smr agent stop control
}

ServiceInstall(){
  LoadEnvFile "$HOME/nodes/.env"

  UNIT_NAME="simplecontainer@.service"
  UNIT_PATH="/etc/systemd/system/${UNIT_NAME}"

  VERSION_SMR=${2:-$(curl -sL https://raw.githubusercontent.com/simplecontainer/smr/refs/heads/main/cmd/smr/version --fail)}
  UNIT_FILE=$(curl -sL https://github.com/simplecontainer/smr/releases/download/smr-$VERSION_SMR/simplecontainer.unit --fail)

  if [[ -z "$UNIT_FILE" ]]; then
    echo "Failed to download systemd unit file"
    exit 1
  fi

  echo "$UNIT_FILE" | sudo tee "$UNIT_PATH" > /dev/null

  sudo systemctl daemon-reload
  sudo systemctl enable simplecontainer@${SUDO_USER:-$USER}
}

ServiceStart(){
  LoadEnvFile "$HOME/nodes/.env"

  Start "$@"
  smr agent events --wait cluster_ready
  smrctl context import $(smr agent export --api $DOMAIN:1443)
  smr agent events
}

ServiceStop(){
  LoadEnvFile "$HOME/nodes/.env"
  Stop "$@"
}

Download(){
  curl --version > /dev/null 2>&1 || echo "Please install curl before proceeding with installing smr!" | exit 1

  ARCH=$(DetectArch)
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

SaveCurrentEnvToFile() {
    local env_file="${1:-.env}"
    local vars_to_save=(
        "NODE"
        "DOMAIN"
        "IP"
        "NODE_ARGS"
        "CLIENT_ARGS"
        "JOIN"
        "SERVICE"
        "PEER"
        "IMAGE"
        "TAG"
    )

    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')

    {
        echo "# SMR Manager Configuration"
        echo "# Generated on: $timestamp"
        echo ""

        for var in "${vars_to_save[@]}"; do
            if [[ -n "${!var}" ]]; then
                echo "${var}=\"${!var}\""
            else
                echo "# ${var}=\"\""
            fi
        done

        echo ""
        echo "# System Info"
        echo "SMR_VERSION=\"$(smr version 2>/dev/null || echo 'not available')\""
        echo "SMRCTL_VERSION=\"$(smrctl version 2>/dev/null || echo 'not available')\""
        echo "TIMESTAMP=\"$timestamp\""

    } > "$env_file"

    echo "Environment variables saved to: $env_file"
}

LoadEnvFile() {
    local env_file="${1:-.env}"

    if [[ -f "$env_file" ]]; then
        echo "Loading environment from: $env_file"
        source "$env_file"
        echo "Environment loaded successfully"
    else
        echo "Error: $env_file not found"
        return 1
    fi
}

DetectArch() {
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
      ServiceStart "$@";;
    "service-stop")
      ServiceStop "$@";;
    "service-install")
      ServiceInstall "$@";;
    *)
      echo "Available commands are: install, start, stop, service-start, service-stop" ;;
esac