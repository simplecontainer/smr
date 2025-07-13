#cloud-config
package_update: true
package_upgrade: false

users:
  - name: node
    shell: /bin/bash
    groups: [sudo, docker]
    sudo: ALL=(ALL) NOPASSWD:ALL
    lock_passwd: false

packages:
  - ca-certificates
  - curl
  - gnupg
  - lsb-release
  - wireguard

write_files:
  - path: /etc/sysctl.d/99-ip-forward.conf
    content: |
      net.ipv4.ip_forward=1
      net.ipv6.conf.all.forwarding=1

runcmd:
  - sysctl --system
  - curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
  - echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" > /etc/apt/sources.list.d/docker.list
  - apt-get update
  - apt-get install -y docker-ce docker-ce-cli containerd.io
  - systemctl enable docker
  - systemctl start docker

  # Install SMR and set up system
  - su - node -c "curl -sL https://raw.githubusercontent.com/simplecontainer/smr/refs/heads/main/scripts/production/smrmgr.sh -o smrmgr"
  - su - node -c "chmod +x smrmgr"
  - su - node -c "sudo mv smrmgr /usr/local/bin"
  - su - node -c "sudo smrmgr install"

  # Get Hetzner metadata
  - export TOKEN={{ .token }}
  - export ACTION={{ .action }}
  - LOCAL_IP=$(curl -s http://169.254.169.254/metadata/local-ipv4)
  - PUBLIC_HOSTNAME=$(curl -s http://169.254.169.254/metadata/hostname)
  - su - node -c "smrmgr start -a $LOCAL_IP -d $PUBLIC_HOSTNAME -s"
  - su - node -c "smrmgr service-install"
  - su - node -c "sudo systemctl set-environment TOKEN="$TOKEN" ACTION="$ACTION"
  - su - node -c "sudo systemctl start simplecontainer@node"
