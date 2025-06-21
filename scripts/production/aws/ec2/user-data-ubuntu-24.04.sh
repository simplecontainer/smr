#!/bin/bash
set -euxo pipefail

curl -fsSL https://get.docker.com | sudo bash
apt-get install -y wireguard

sysctl -w net.ipv4.ip_forward=1
sysctl -w net.ipv6.conf.all.forwarding=1
echo "net.ipv4.ip_forward=1" >> /etc/sysctl.conf
echo "net.ipv6.conf.all.forwarding=1" >> /etc/sysctl.conf

useradd -m -s /bin/bash simplecontainer
usermod -aG sudo simplecontainer
usermod -aG docker simplecontainer
echo "simplecontainer ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers.d/90-cloud-init-users

su - simplecontainer -c "bash -s" <<'EOF'
set -euxo pipefail
curl -sL https://raw.githubusercontent.com/simplecontainer/smr/refs/heads/main/scripts/production/smrmgr.sh -o smrmgr
chmod +x smrmgr
sudo mv smrmgr /usr/local/bin
sudo smrmgr install
smrmgr start
EOF
