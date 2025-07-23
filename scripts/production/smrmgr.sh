#!/bin/bash
#
# Smrmgr - management script for simplecontainer
# Author: Adnan Selimovic (adnn.selimovic@gmail.com)
# Version: 0.0.1
#

set -euo pipefail

# ============================================================================
# CONSTANTS AND CONFIGURATION
# ============================================================================

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_NAME="$(basename "$0")"
readonly SCRIPT_VERSION="2.0.0"

# Default configuration
readonly DEFAULT_REGISTRY="http://api.simplecontainer.io"
readonly DEFAULT_IMAGE="quay.io/simplecontainer/smr"
readonly DEFAULT_NODE_NAME="simplecontainer-node-1"
readonly DEFAULT_DOMAIN="localhost"
readonly DEFAULT_NODE_ARGS="--listen 0.0.0.0:1443"
readonly DEFAULT_CLIENT_ARGS="--port.control 0.0.0.0:1443 --port.overlay 0.0.0.0:9212"
readonly NODES_DIR="$HOME/nodes"
readonly ENV_FILE="$NODES_DIR/.env"

# URLs
readonly VERSION_URL="https://raw.githubusercontent.com/simplecontainer/smr/refs/heads/main/cmd/smr/version"
readonly SMRCTL_VERSION_URL="https://raw.githubusercontent.com/simplecontainer/smr/refs/heads/main/cmd/smrctl/version"
readonly SYSTEMD_UNIT_NAME="simplecontainer@.service"
readonly SYSTEMD_UNIT_PATH="/etc/systemd/system/${SYSTEMD_UNIT_NAME}"

# ============================================================================
# GLOBAL VARIABLES
# ============================================================================

# Node configuration
declare -g NODE_NAME="${NODE_NAME:-}"
declare -g DOMAIN="${DOMAIN:-}"
declare -g IP_ADDRESS="${IP_ADDRESS:-}"
declare -g NODE_ARGS="${NODE_ARGS:-}"
declare -g CLIENT_ARGS="${CLIENT_ARGS:-}"
declare -g JOIN_CLUSTER="${JOIN_CLUSTER:-false}"
declare -g PEER_ADDRESS="${PEER_ADDRESS:-}"
declare -g DOCKER_IMAGE="${DOCKER_IMAGE:-}"
declare -g DOCKER_TAG="${DOCKER_TAG:-}"
declare -g INSTALL_SERVICE="${INSTALL_SERVICE:-false}"
declare -g TOKEN="${TOKEN:-}"
declare -g ACTION="${ACTION:-}"

# ============================================================================
# UTILITY FUNCTIONS
# ============================================================================

log() {
    local level="$1"
    shift
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] [$level] $*" >&2
}

log_info() {
    log "INFO" "$@"
}

log_error() {
    log "ERROR" "$@"
}

log_debug() {
    if [[ "${DEBUG:-false}" == "true" ]]; then
        log "DEBUG" "$@"
    fi
}

die() {
    log_error "$@"
    exit 1
}

check_dependencies() {
    local missing_deps=()

    command -v curl >/dev/null 2>&1 || missing_deps+=("curl")
    command -v docker >/dev/null 2>&1 || missing_deps+=("docker")
    command -v smr >/dev/null 2>&1 || missing_deps+=("smr")
    command -v smrctl >/dev/null 2>&1 || missing_deps+=("smrctl")

    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        die "Missing dependencies: ${missing_deps[*]}. Please install them first."
    fi
}

validate_input() {
    local input="$1"
    local pattern="$2"
    local description="$3"

    if [[ ! $input =~ $pattern ]]; then
        die "Invalid $description: $input"
    fi
}

create_directory() {
    local dir="$1"
    if [[ ! -d "$dir" ]]; then
        mkdir -p "$dir" || die "Failed to create directory: $dir"
        log_info "Created directory: $dir"
    fi
}

fetch_latest_version() {
    local url="$1"
    local version
    version=$(curl -sL "$url" --fail) || die "Failed to fetch version from $url"
    echo "$version"
}

# ============================================================================
# ARCHITECTURE DETECTION
# ============================================================================

detect_architecture() {
    local arch=""

    if command -v uname >/dev/null 2>&1; then
        arch="$(uname -m 2>/dev/null)"
    elif command -v dpkg >/dev/null 2>&1; then
        arch="$(dpkg --print-architecture 2>/dev/null)"
    elif command -v arch >/dev/null 2>&1; then
        arch="$(arch 2>/dev/null)"
    fi

    case "$arch" in
        x86_64|amd64)
            echo "amd64"
            ;;
        aarch64|arm64)
            echo "arm64"
            ;;
        armv8*|armv7*|armv6*|armhf)
            echo "arm64"
            ;;
        *)
            die "Unknown or unsupported architecture: '$arch'"
            ;;
    esac
}

# ============================================================================
# CONFIGURATION MANAGEMENT
# ============================================================================

initialize_defaults() {
    NODE_NAME="${NODE_NAME:-$DEFAULT_NODE_NAME}"
    DOMAIN="${DOMAIN:-$DEFAULT_DOMAIN}"
    NODE_ARGS="${NODE_ARGS:-$DEFAULT_NODE_ARGS}"
    CLIENT_ARGS="${CLIENT_ARGS:-$DEFAULT_CLIENT_ARGS}"
    DOCKER_IMAGE="${DOCKER_IMAGE:-$DEFAULT_IMAGE}"
    DOCKER_TAG="${DOCKER_TAG:-$(fetch_latest_version "$VERSION_URL")}"
}

parse_arguments() {
    local OPTIND
    while getopts "n:d:a:c:i:t:jp:sT:A:h" option; do
        case $option in
            n) NODE_NAME="$OPTARG" ;;
            d) DOMAIN="$OPTARG" ;;
            a) IP_ADDRESS="$OPTARG" ;;
            c) CLIENT_ARGS="$OPTARG" ;;
            i) DOCKER_IMAGE="$OPTARG" ;;
            t) DOCKER_TAG="$OPTARG" ;;
            j) JOIN_CLUSTER="true" ;;
            p) PEER_ADDRESS="$OPTARG" ;;
            s) INSTALL_SERVICE="true" ;;
            T) TOKEN="$OPTARG" ;;
            A) ACTION="$OPTARG" ;;
            h) show_help && exit 0 ;;
            \?) die "Invalid option: -$OPTARG" ;;
            :) die "Option -$OPTARG requires an argument." ;;
        esac
    done
}

validate_configuration() {
    [[ -n "$NODE_NAME" ]] || die "Node name is required"
    [[ -n "$DOMAIN" ]] || die "Domain is required"
    [[ -n "$DOCKER_IMAGE" ]] || die "Docker image is required"
    [[ -n "$DOCKER_TAG" ]] || die "Docker tag is required"

    validate_input "$DOMAIN" '^[a-zA-Z0-9.-]+$' "domain"

    if [[ -n "$IP_ADDRESS" ]]; then
        validate_input "$IP_ADDRESS" '^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$' "IP address"
    fi

    if [[ "$JOIN_CLUSTER" == "true" && -z "$PEER_ADDRESS" ]]; then
        die "Peer address is required when joining a cluster"
    fi
}

build_node_arguments() {
    local args="--image ${DOCKER_IMAGE} --tag ${DOCKER_TAG} --node ${NODE_NAME} ${NODE_ARGS}"

    [[ -n "$DOMAIN" ]] && args+=" --domain ${DOMAIN}"
    [[ -n "$IP_ADDRESS" ]] && args+=" --ip ${IP_ADDRESS}"
    [[ "$JOIN_CLUSTER" == "true" && -n "$PEER_ADDRESS" ]] && args+=" --join --peer ${PEER_ADDRESS}"

    echo "$args"
}

# ============================================================================
# ENVIRONMENT MANAGEMENT
# ============================================================================

save_environment() {
    local env_file="${1:-$ENV_FILE}"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')

    create_directory "$(dirname "$env_file")"

    cat > "$env_file" << EOF
# SMR Manager Configuration
# Generated on: $timestamp
# Script version: $SCRIPT_VERSION

NODE_NAME="$NODE_NAME"
DOMAIN="$DOMAIN"
IP_ADDRESS="$IP_ADDRESS"
NODE_ARGS="$NODE_ARGS"
CLIENT_ARGS="$CLIENT_ARGS"
JOIN_CLUSTER="$JOIN_CLUSTER"
PEER_ADDRESS="$PEER_ADDRESS"
DOCKER_IMAGE="$DOCKER_IMAGE"
DOCKER_TAG="$DOCKER_TAG"
INSTALL_SERVICE="$INSTALL_SERVICE"

# System Info
SMR_VERSION="$(smr version 2>/dev/null || echo 'not available')"
SMRCTL_VERSION="$(smrctl version 2>/dev/null || echo 'not available')"
TIMESTAMP="$timestamp"
EOF

    log_info "Environment saved to: $env_file"
}

load_environment() {
    local env_file="${1:-$ENV_FILE}"

    if [[ -f "$env_file" ]]; then
        log_info "Loading environment from: $env_file"
        # shellcheck source=/dev/null
        source "$env_file"
        log_info "Environment loaded successfully"
    else
        log_error "Environment file not found: $env_file"
        return 1
    fi
}

# ============================================================================
# NODE MANAGEMENT
# ============================================================================

create_node() {
    local node_args
    node_args=$(build_node_arguments)

    log_info "Creating node with arguments: $node_args"

    if ! smr node create --node "$NODE_NAME" $node_args $CLIENT_ARGS; then
        die "Failed to create node configuration"
    fi

    # Create log directories
    local log_dir="$NODES_DIR/$NODE_NAME/logs"
    create_directory "$log_dir"

    touch "$log_dir/cluster.log" || die "Failed to create cluster log file"
    touch "$log_dir/control.log" || die "Failed to create control log file"

    log_info "Node '$NODE_NAME' created successfully"
}

start_node() {
    log_info "Starting node: $NODE_NAME"

    smr node start --node "$NODE_NAME" -y || die "Failed to start node"

    local raft_url
    if [[ "$DOMAIN" == "localhost" ]]; then
        local container_ip
        container_ip=$(docker inspect -f '{{.NetworkSettings.Networks.bridge.IPAddress}}' "$NODE_NAME")
        raft_url="https://${container_ip}:9212"
    else
        raft_url="https://${DOMAIN}:9212"
    fi

    log_info "Starting SMR agent with RAFT URL: $raft_url"

    # Start cluster agent
    sudo nohup smr agent start --node "$NODE_NAME" --raft "$raft_url" \
        </dev/null 2>&1 | stdbuf -o0 grep "" > "$NODES_DIR/$NODE_NAME/logs/cluster.log" &

    # Start control agent
    nohup smr agent control --node "$NODE_NAME" \
        </dev/null 2>&1 | stdbuf -o0 grep "" > "$NODES_DIR/$NODE_NAME/logs/control.log" &

    log_info "Node '$NODE_NAME' started successfully"
}

stop_node() {
    log_info "Stopping node: $NODE_NAME"

    smr agent drain --node "$NODE_NAME" || log_error "Failed to drain node"
    smr agent events --node "$NODE_NAME" --wait drain_success || log_error "Failed to wait for drain completion"
    smr node clean --node "$NODE_NAME" || log_error "Failed to clean node"

    sudo smr agent stop agent || log_error "Failed to stop agent"
    smr agent stop control || log_error "Failed to stop control"

    log_info "Node '$NODE_NAME' stopped successfully"
}

# ============================================================================
# CLUSTER MANAGEMENT
# ============================================================================

wait_for_cluster_ready() {
    log_info "Waiting for cluster to be ready..."
    smr agent events --wait cluster_started --node "$NODE_NAME" || die "Cluster failed to start"
    log_info "Cluster is ready"
}

show_cluster_info() {
    echo
    echo "================================================================================================"
    echo "Node Information"
    echo "================================================================================================"
    echo "Agent name:           $NODE_NAME"
    echo "Domain:               $DOMAIN"
    echo "Image:                $DOCKER_IMAGE"
    echo "Tag:                  $DOCKER_TAG"
    echo "IP Address:           ${IP_ADDRESS:-'auto-detected'}"
    echo "Join cluster:         $JOIN_CLUSTER"
    echo "Peer address:         ${PEER_ADDRESS:-'N/A'}"
    echo "Service install:      $INSTALL_SERVICE"

    # Only show TOKEN and ACTION if they're set (first-time startup only)
    if [[ -n "$TOKEN" ]]; then
        echo "Token:                ${TOKEN:0:8}..."
    fi
    if [[ -n "$ACTION" ]]; then
        echo "Action:               $ACTION"
    fi

    echo "smr version:          $(smr version 2>/dev/null || echo 'not available')"
    echo "smrctl version:       $(smrctl version 2>/dev/null || echo 'not available')"
    echo "================================================================================================"
    echo
    echo "Log files:"
    echo "  Cluster: tail -f $NODES_DIR/$NODE_NAME/logs/cluster.log"
    echo "  Control: tail -f $NODES_DIR/$NODE_NAME/logs/control.log"
    echo "================================================================================================"
}

# ============================================================================
# SERVICE MANAGEMENT
# ============================================================================

install_systemd_service() {
    log_info "Installing systemd service..."

    load_environment || die "Failed to load environment for service installation"

    local version_smr
    version_smr=$(fetch_latest_version "$VERSION_URL")

    local unit_file
    unit_file=$(curl -sL "https://github.com/simplecontainer/smr/releases/download/smr-$version_smr/simplecontainer.unit" --fail) || \
        die "Failed to download systemd unit file"

    echo "$unit_file" | sudo tee "$SYSTEMD_UNIT_PATH" > /dev/null || \
        die "Failed to write systemd unit file"

    sudo systemctl daemon-reload || die "Failed to reload systemd daemon"
    sudo systemctl enable "simplecontainer@${SUDO_USER:-$USER}" || die "Failed to enable service"

    log_info "Systemd service installed successfully"
}

service_start() {
    log_info "Starting systemd service..."

    load_environment || die "Failed to load environment"

    if [[ -n "$ACTION" ]]; then
        log_info "Performing action: $ACTION"

        # TOKEN and ACTION are available here from parse_arguments only on the bootstrap-start!
        # Use them for first-time startup logic
        if [[ -n "$TOKEN" ]]; then
            log_info "Using authentication token for service startup"
        else
          die "Token is needed when providing action."
        fi

        case "$ACTION" in
            standalone)
                start_node
                wait_for_cluster_ready

                smrctl context import --y $(smr agent export --node "$NODE_NAME") || \
                    log_error "Failed to import context"

                smrctl context export active --upload --token $TOKEN --api "$DOMAIN:1443" --registry "$DEFAULT_REGISTRY" || \
                    log_error "Failed to export context to the registry"
                ;;
            cluster-leader)
                start_node
                wait_for_cluster_ready

                smrctl context import --y $(smr agent export --node "$NODE_NAME") || \
                    log_error "Failed to import context"

                smrctl context export active --upload --token $TOKEN --api "$DOMAIN:1443" --registry "$DEFAULT_REGISTRY" || \
                    log_error "Failed to export context to the registry"
                ;;
            cluster-join)
                smrctl context import --download --token $TOKEN  --registry "$DEFAULT_REGISTRY" --y || \
                    log_error "Failed to import context from the registry to the smrctl"

                smr agent import --node "$NODE_NAME" --y $(smrctl context export active) || \
                    log_error "Failed to import context from the registry to the smr agent"

                start_node
                wait_for_cluster_ready

                smrctl context import $(smr agent export --api "$DOMAIN:1443") || \
                    log_error "Failed to import context"
                ;;
            *)
                # Default case (catch-all)
                die "Invalid action selected!"
                ;;
        esac

        sudo systemctl unset-environment TOKEN ACTION
    else
      start_node
      wait_for_cluster_ready

      smrctl context import $(smr agent export --node "$NODE_NAME" --api "$DOMAIN:1443") || \
          log_error "Failed to import context"
    fi

    log_info "Service started successfully - now listening events and outputting in journal!"
    smr agent events --node $NODE_NAME || log_error "Failed to show agent events"
}

service_stop() {
    log_info "Stopping systemd service..."

    load_environment || die "Failed to load environment"
    stop_node

    log_info "Service stopped successfully"
}

service_install() {
    install_systemd_service "$@"
}

# ============================================================================
# INSTALLATION MANAGEMENT
# ============================================================================

download_binaries() {
    local version_smr="${1:-$(fetch_latest_version "$VERSION_URL")}"
    local version_ctl="${2:-$(fetch_latest_version "$SMRCTL_VERSION_URL")}"
    local arch
    arch=$(detect_architecture)
    local platform="linux-${arch}"

    log_info "Downloading smr:$version_smr and smrctl:$version_ctl for platform: $platform"

    # Download smr
    local smr_url="https://github.com/simplecontainer/smr/releases/download/smr-$version_smr/smr-$platform"
    log_info "Downloading: $smr_url"

    curl -Lo smr "$smr_url" --fail || die "Failed to download smr binary"
    chmod +x smr || die "Failed to make smr executable"

    # Verify smr binary
    if ! ./smr --help >/dev/null 2>&1 && ! ./smr --version >/dev/null 2>&1; then
        die "Downloaded smr binary is not executable or cannot run"
    fi

    # Download smrctl
    local smrctl_url="https://github.com/simplecontainer/smr/releases/download/smrctl-$version_ctl/smrctl-$platform"
    log_info "Downloading: $smrctl_url"

    curl -Lo smrctl "$smrctl_url" --fail || die "Failed to download smrctl binary"
    chmod +x smrctl || die "Failed to make smrctl executable"

    # Verify smrctl binary
    if ! ./smrctl --help >/dev/null 2>&1 && ! ./smrctl --version >/dev/null 2>&1; then
        die "Downloaded smrctl binary is not executable or cannot run"
    fi

    # Install binaries
    sudo mv smr /usr/local/bin/smr || die "Failed to install smr"
    sudo mv smrctl /usr/local/bin/smrctl || die "Failed to install smrctl"

    log_info "Binaries installed successfully to /usr/local/bin"
}

# ============================================================================
# MAIN FUNCTIONS
# ============================================================================

cmd_start() {
    parse_arguments "$@"
    initialize_defaults
    validate_configuration
    check_dependencies

    show_cluster_info

    create_node

    if [[ "$INSTALL_SERVICE" == "false" ]]; then
        start_node
        wait_for_cluster_ready

        echo
        echo "Node started successfully!"
        echo "Log files are available at:"
        echo "  Cluster: tail -f $NODES_DIR/$NODE_NAME/logs/cluster.log"
        echo "  Control: tail -f $NODES_DIR/$NODE_NAME/logs/control.log"
    fi

    save_environment
}

cmd_stop() {
    parse_arguments "$@"
    load_environment || die "Failed to load environment"
    stop_node
}

cmd_install() {
    local version_smr="${1:-}"
    local version_ctl="${2:-}"

    check_dependencies() {
        command -v curl >/dev/null 2>&1 || die "Please install curl before proceeding"
    }

    check_dependencies
    download_binaries "$version_smr" "$version_ctl"
}

cmd_service_install() {
    install_systemd_service
}

cmd_service_start() {
  parse_arguments "$@"
  service_start
}

cmd_service_stop() {
    service_stop
}

show_help() {
    cat << EOF
SMR Manager - Version $SCRIPT_VERSION
A modular management script for simplecontainer

USAGE:
    $SCRIPT_NAME <command> [options]

COMMANDS:
    install                 Download and install smr and smrctl binaries
    start                   Start a new node or join existing cluster
    stop                    Stop the current node
    service-install         Install systemd service
    service-start           Start the systemd service
    service-stop            Stop the systemd service

START OPTIONS:
    -n <node>              Set node name (default: $DEFAULT_NODE_NAME)
    -d <domain>            Set domain (default: $DEFAULT_DOMAIN)
    -a <ip>                Set IP address (auto-detected if not specified)
    -c <args>              Set additional client arguments
    -i <image>             Set Docker image (default: $DEFAULT_IMAGE)
    -t <tag>               Set Docker image tag (default: latest from repo)
    -j                     Join an existing cluster
    -p <peer>              Set peer address (required when joining)
    -s                     Install as systemd service
    -T <token>             Set authentication token (first-time startup only)
    -A <action>            Set action to perform (first-time startup only)
    -h                     Show this help message

EXAMPLES:
    # Install binaries
    $SCRIPT_NAME install

    # Start a new cluster
    $SCRIPT_NAME start -n node-1 -d mydomain.com -a 10.0.0.1

    # Join an existing cluster
    $SCRIPT_NAME start -n node-2 -j -p 10.0.0.1

    # Start with custom image and service installation
    $SCRIPT_NAME start -n node-1 -i myrepo/myimage -t latest -s

    # Start with token and action (first-time startup)
    $SCRIPT_NAME start -n node-1 -T mytoken123 -A deploy

    # Install systemd service
    $SCRIPT_NAME service-install

    # Start service with token/action (first-time startup used with app.simplecontainer.io only)
    $SCRIPT_NAME service-start -T mytoken123 -A bootstrap

    # Subsequent service starts (no token/action needed)
    $SCRIPT_NAME service-start

ENVIRONMENT:
    DEBUG=true             Enable debug logging
    NODES_DIR              Override nodes directory (default: ~/nodes)

For more information, visit: https://github.com/simplecontainer/smr
EOF
}

# ============================================================================
# MAIN ENTRY POINT
# ============================================================================

main() {
    local command="${1:-help}"
    shift || true

    case "$command" in
        install)
            cmd_install "$@"
            ;;
        start)
            cmd_start "$@"
            ;;
        stop)
            cmd_stop "$@"
            ;;
        service-install)
            cmd_service_install "$@"
            ;;
        service-start)
            cmd_service_start "$@"
            ;;
        service-stop)
            cmd_service_stop "$@"
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            echo "Unknown command: $command"
            echo "Run '$SCRIPT_NAME help' for usage information"
            exit 1
            ;;
    esac
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi