#!/bin/bash

#
# Notice:
# Script can be outdated and is maintained on the best-effort basis.
# Last update: Jun 23. 2025 (Please update here if you do any work on keeping it up to date.)
#
# This script does next:
# - install docker-ce
# - install wireguard
# - install smr, smrctl, and smrmgr.sh
# - setup simplecontainer to work on boot using services

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
    centos|rhel|fedora)
        yum install -y yum-utils wireguard-tools curl --skip-broken
        yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
        yum install -y docker-ce docker-ce-cli containerd.io || dnf install -y docker-ce
        systemctl enable docker
        systemctl start docker
        SUDO_GROUP="wheel"
        ;;
    amzn)
        if grep -q "2023" /etc/os-release; then
            # Amazon Linux 2023
            dnf update -y
            dnf install -y dnf-plugins-core wireguard-tools curl --skip-broken

            dnf config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
            sed -i 's/$releasever/9/g' /etc/yum.repos.d/docker-ce.repo
            dnf -y install docker-ce docker-ce-cli containerd.io docker-buildx-plugin
        else
            # Amazon Linux 2
            amazon-linux-extras enable docker
            yum clean metadata
            yum install -y docker wireguard-tools curl --skip-broken
        fi
        systemctl enable docker
        systemctl start docker
        SUDO_GROUP="wheel"
        ;;
    sles|opensuse*|suse)
        zypper refresh
        zypper install -y docker wireguard-tools sudo curl ca-certificates
        systemctl enable docker
        systemctl start docker
        SUDO_GROUP="wheel"
        ;;
    *)
        echo "Unsupported distro: $DISTRO"
        exit 1
        ;;
esac

sysctl -w net.ipv4.ip_forward=1
sysctl -w net.ipv6.conf.all.forwarding=1
echo "net.ipv4.ip_forward=1" >> /etc/sysctl.conf
echo "net.ipv6.conf.all.forwarding=1" >> /etc/sysctl.conf

useradd -m -s /bin/bash node
usermod -g docker node
usermod -aG $SUDO_GROUP node
echo "node ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers.d/90-cloud-init-users

DEFAULT_USER=$(for u in ec2-user ubuntu admin centos; do id $u &>/dev/null && echo $u && break; done)
cp -r /home/$DEFAULT_USER/.ssh /home/node/ && chown -R node /home/node/.ssh

su - node -c "bash -s" <<'EOF'
set -euxo pipefail
curl -sL https://raw.githubusercontent.com/simplecontainer/smr/refs/heads/main/scripts/production/smrmgr.sh -o smrmgr
chmod +x smrmgr
sudo mv smrmgr /usr/local/bin
sudo smrmgr install

TOKEN=$(curl -s -X PUT "http://169.254.169.254/latest/api/token" -H "X-aws-ec2-metadata-token-ttl-seconds: 21600")
LOCAL_IP=$(curl -H "X-aws-ec2-metadata-token: $TOKEN" http://169.254.169.254/latest/meta-data/local-ipv4)
PUBLIC_HOSTNAME=$(curl -H "X-aws-ec2-metadata-token: $TOKEN" http://169.254.169.254/latest/meta-data/public-hostname)

smrmgr start -a $LOCAL_IP -d $PUBLIC_HOSTNAME -s
smrmgr service-install

sudo systemctl start simplecontainer@${SUDO_USER:-$USER}
EOF