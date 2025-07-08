#!/bin/bash

#
# Maintainer note:
# Last updated for Hetzner: Jul 8, 2025
#

# This script:
# - Installs Docker CE
# - Installs WireGuard
# - Installs SMR, smrctl, and smrmgr.sh
# - Sets up systemd service for simplecontainer

set -euxo pipefail

# Detect distro
DISTRO=""
if [ -f /etc/os-release ]; then
    . /etc/os-release
    DISTRO=$ID
else
    echo "Cannot detect Linux distro"
    exit 1
fi

# Install Docker and WireGuard
case "$DISTRO" in
    ubuntu|debian)
        export DEBIAN_FRONTEND=noninteractive
        apt-get update
        apt-get install -y ca-certificates curl gnupg lsb-release wireguard
        curl -fsSL https://download.docker.com/linux/$DISTRO/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
        echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/$DISTRO $(lsb_release -cs) stable" > /etc/apt/sources.list.d/docker.list
        apt-get update
        apt-get install -y docker-ce docker-ce-cli containerd.io
        systemctl enable docker
        systemctl start docker
        SUDO_GROUP="sudo"
        ;;
    *)
        echo "Unsupported distro: $DISTRO"
        exit 1
        ;;
esac

# Enable IP forwarding
sysctl -w net.ipv4.ip_forward=1
sysctl -w net.ipv6.conf.all.forwarding=1
echo "net.ipv4.ip_forward=1" >> /etc/sysctl.conf
echo "net.ipv6.conf.all.forwarding=1" >> /etc/sysctl.conf

# Create 'node' user and add to docker/sudo
useradd -m -s /bin/bash node
usermod -aG docker node
usermod -aG $SUDO_GROUP node
echo "node ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers.d/90-cloud-init-users

# Setup SMR as 'node' user
su - node -c "bash -s" <<'EOF'
set -euxo pipefail

# Install smrmgr
curl -sL https://raw.githubusercontent.com/simplecontainer/smr/refs/heads/main/scripts/production/smrmgr.sh -o smrmgr
chmod +x smrmgr
sudo mv smrmgr /usr/local/bin
sudo smrmgr install

# Get local IP and public hostname from Hetzner metadata
METADATA_BASE="http://169.254.169.254/metadata"
LOCAL_IP=$(curl -s "${METADATA_BASE}/local-ipv4")
PUBLIC_HOSTNAME=$(curl -s "${METADATA_BASE}/hostname")

# Start smr with IPs
smrmgr start -a $LOCAL_IP -d $PUBLIC_HOSTNAME -s
smrmgr service-install

# Start container service
sudo systemctl start simplecontainer@${SUDO_USER:-$USER}
EOF
