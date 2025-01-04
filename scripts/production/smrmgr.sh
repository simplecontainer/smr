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

Start(){
  AGENT=""
  DOMAIN=""
  IP=""
  NODE_DOMAIN=""
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
           NODE_DOMAIN=$OPTARG;;
        p) # Set node port
           NODE_PORT=$OPTARG;;
        r) # Set repository
           REPOSITORY=$OPTARG;;
        x) # Set client additional args
           CLIENT_ARGS=$OPTARG;;
        *) # Invalid option
          echo "Invalid option";;
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

  echo "Agent name: $AGENT"
  echo "Node URL: $NODE_DOMAIN"
  echo "Domain: $DOMAIN"
  echo "Additional arguments: $CLIENT_ARGS"
  echo "Join: $JOIN"

  if [[ ${AGENT} != "" ]]; then
    if [[ ${MODE} == "cluster" ]]; then
      smr node run --image "${REPOSITORY}" --tag "${TAG}" --args="create --agent ${AGENT} --port ${CONTROL_PLANE} --domains ${DOMAIN} --ips ${IP}" --agent "${AGENT}" $CLIENT_ARGS --wait

      if [[ ${?} != 0 ]]; then
        echo "Smr returned non-zero exit code - check the logs of the node controller container"
        exit
      fi

      smr node run --image "${REPOSITORY}" --tag "${TAG}" --args="start" $CLIENT_ARGS --agent "${AGENT}"

      if [[ $NODE_DOMAIN == "localhost" ]]; then
        NODE_DOMAIN="https://$(docker inspect -f '{{.NetworkSettings.Networks.bridge.IPAddress}}' $AGENT):${NODE_PORT}"
      else
        NODE_DOMAIN="https://${NODE_DOMAIN}:${NODE_PORT}"
      fi

      while :
      do
      	if smr context connect "${CONN_STRING}" "${HOME}/.ssh/simplecontainer/${AGENT}.pem" --context "${AGENT}" --wait --y; then
      	  break
      	else
      	  echo "Failed to connect to siplecontainer, trying again in 1 second"
          sleep 1
      	fi
      done

      if [[ ${JOIN} == "" ]]; then
        sudo nohup smr node cluster start --node "${NODE_DOMAIN}" 2>&1 | dd of=~/smr/smr/logs/$AGENT-cluster.log &>/dev/null &
      else
        sudo nohup smr node cluster start --node "${NODE_DOMAIN}" --join "https://${JOIN}" 2>&1 | dd of=~/smr/smr/logs/$AGENT-cluster-join.log &>/dev/null &
      fi

      echo "The simplecontainer is started in cluster mode."
    else
      smr node run --image "${REPOSITORY}" --tag "${TAG}" --args="create --agent ${AGENT} --domain ${DOMAIN} --ip ${IP}" --agent "${AGENT}" --wait
      smr node run --image "${REPOSITORY}" --tag "${TAG}" --args="start" $CLIENT_ARGS "${CONTROL_PLANE}" --agent "${AGENT}"

      sudo nohup smr node cluster start --node "${NODE_DOMAIN}" 2>&1 | dd of=~/smr/smr/logs/$AGENT-cluster.log &>/dev/null &

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
