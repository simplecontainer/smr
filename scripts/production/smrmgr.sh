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
  CLIENT_ARGS="--overlayport 0.0.0.0:9212"
  MODE="cluster"
  JOIN=""
  CONTROL_PLANE="0.0.0.0:1443"
  REPOSITORY="quay.io/simplecontainer/smr"
  TAG=$(curl -sL https://raw.githubusercontent.com/simplecontainer/smr/main/version)
  RESTART=false
  UPGRADE=false
  ALLYES=false

  while getopts ":a:c:d:e:h:i:j:m:n:p:r:s:t:u:x:y:z:" option; do
    case $option in
      a) # Set agent
         NODE=$OPTARG;;
      c) # Control plane client connect string
         CONN_STRING=$OPTARG;;
      d) # Set domain
         DOMAIN=$OPTARG;;
      e) # Expose control plane
         CONTROL_PLANE=$OPTARG;;
      h) # Display help
         HelpStart && exit;;
      i) # Set ip
         IP=$OPTARG;;
      j) #Set join
         JOIN=$OPTARG;;
      m) #Set mode
         MODE=$OPTARG;;
      n) # Set Node URL
         NODE_DOMAIN=$OPTARG;;
      p) # Set node port
         NODE_PORT=$OPTARG;;
      r) # Set repository
         REPOSITORY=$OPTARG;;
      t) # Set tag
         TAG=$OPTARG;;
      u) # Set upgrade true/false
         UPGRADE=$OPTARG;;
      s) # Set restart true/false
         RESTART=$OPTARG;;
      x) # Set client additional args
         CLIENT_ARGS=$OPTARG;;
      y) #SSet all yes answer
         ALLYES=$OPTARG;;
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

  echo "Agent name: $NODE"
  echo "Node: $NODE_DOMAIN"
  echo "Additional arguments: $CLIENT_ARGS"
  echo "Restart: $RESTART"
  echo "Upgrade: $UPGRADE"

  if [[ $JOIN != "" ]]; then
    echo "Join: $JOIN"
  fi

  smr version
  touch ~/smr/smr/logs/flannel-${NODE}.log

  if ! dpkg -s jq &>/dev/null; then
    echo 'please install jq manually'
    exit 1
  fi

  if [[ "${UPGRADE}" == true || "${RESTART}" == true ]]; then
    read -p "This action will stop simplecontainer node and restart it from config file [Yy/Nn]? " -n 1 -r
    echo

    if [[ $REPLY =~ ^[Yy]$ || $ALLYES == "true" ]]; then
        IMAGETAG=$(smr cli inspect --name smr-agent-1 | jq '.Config.Image' | tr -d /\"//)
        arrIN=(${IMAGETAG//:/ })

        smr node stop --name "${NODE}" $CLIENT_ARGS --wait

        if [[ $RESTART == "true" ]]; then
          smr node run --image "${arrIN[0]}" --tag "${arrIN[1]}" --args="start --restore true" $CLIENT_ARGS --name "${NODE}"
        else
          smr node run --image "${REPOSITORY}" --tag "${TAG}" --args="start --restore true" $CLIENT_ARGS --name "${NODE}"
        fi

        while :
        do
          if smr context connect "${CONN_STRING}" "${HOME}/.ssh/simplecontainer/${NODE}.pem" --context "${NODE}" --wait --y; then
            break
          else
            echo "Failed to connect to siplecontainer, trying again in 1 second"
            sleep 1
          fi
        done

        sudo nohup smr node cluster restore </dev/null 2>&1 | stdbuf -o0 grep "" > ~/smr/smr/logs/flannel-${NODE}.log &
    fi
  else
    if [[ ${NODE} != "" ]]; then
      if [[ ${MODE} == "cluster" ]]; then
        smr node run --image "${REPOSITORY}" --tag "${TAG}" --args="create --name ${NODE} --port ${CONTROL_PLANE} --domains ${DOMAIN} --ips ${IP}" --name "${NODE}" $CLIENT_ARGS  --wait

        if [[ ${?} != 0 ]]; then
          echo "Simplecontainer returned non-zero exit code - check the logs of the node controller container"
          exit
        fi

        smr node run --image "${REPOSITORY}" --tag "${TAG}" --args="start" $CLIENT_ARGS --name "${NODE}"

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

        if [[ ${JOIN} == "" ]]; then
          sudo nohup smr node cluster start --node "${NODE_DOMAIN}" </dev/null 2>&1 | stdbuf -o0 grep "" > ~/smr/smr/logs/flannel-${NODE}.log &
        else
          sudo nohup smr node cluster start --node "${NODE_DOMAIN}" --join "https://${JOIN}" </dev/null 2>&1 | stdbuf -o0 grep "" > ~/smr/smr/logs/flannel-${NODE}.log &
        fi

        echo "The simplecontainer is started in cluster mode."
      else
        smr node run --image "${REPOSITORY}" --tag "${TAG}" --args="create --name ${NODE} --domain ${DOMAIN} --ip ${IP}" --name "${NODE}" --wait
        smr node run --image "${REPOSITORY}" --tag "${TAG}" --args="start" $CLIENT_ARGS "${CONTROL_PLANE}" --name "${NODE}"

        sudo nohup smr node cluster start --name "${NODE}" --node "${NODE_DOMAIN}" </dev/null 2>&1 | stdbuf -o0 grep "" > ~/smr/smr/logs/flannel-${NODE}.log &

        echo "The simplecontainer is started in single mode."
      fi
    else
      HelpStart
    fi
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
      Manager "$@";;
    "restart")
      shift
      Manager "-s" "true" "$@";;
    "upgrade")
      shift
      Manager "-u" "true" "$@";;
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
