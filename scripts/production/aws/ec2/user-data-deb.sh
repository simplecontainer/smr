#!/bin/bash

#
# Notice:
# Script can be outdated and is maintained on the best-effort basis.
#
# This script does next:
# - install docker-ce
# - install wireguard
# - install smr, smrctl, and smrmgr.sh
# - setup simplecontainer to work on boot using services

set -euxo pipefail

curl -fsSL https://get.docker.com | sudo bash
apt-get install -y wireguard

sysctl -w net.ipv4.ip_forward=1
sysctl -w net.ipv6.conf.all.forwarding=1
echo "net.ipv4.ip_forward=1" >> /etc/sysctl.conf
echo "net.ipv6.conf.all.forwarding=1" >> /etc/sysctl.conf

useradd -m -s /bin/bash node
usermod -g docker node
usermod -aG sudo node
echo "node ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers.d/90-cloud-init-users

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
