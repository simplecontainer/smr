![simplecontainer manager](.github/resources/repository.jpg)

Quick start
===========

> [!WARNING]
> The project is not stable yet. Releases and major changes are introduced often. 

This is a quick start tutorial for getting a simple container up and running.

## Introduction
A simple container manager is designed to ease life for the developers and DevOps engineers running containers on Docker.

Introducing objects which can be defined as YAML definition and sent to the simplecontainer manager to produce Docker container via reconciliation:

- Containers
- Container
- Configuration
- Resource
- Gitops
- CertKey
- HttpAuth

These objects let you deploy container on local/remote Docker daemon. The simplecontainer introduces next features:

- Single Docker daemon / Cluster of Docker daemons
- Overlay networking using flannel
- Integrated DNS server isolated from Docker daemon
- GitOps: deploy objects from the Git repositories using GitOps approach
- Replication of containers in cluster of Docker daemons
- Reconciliation and tracking the lifecycle of the Docker containers
- CLI to interact with the simplecontainer
- Fast learning curve - simplicity and deterministic behavior
- Reliable dependency ordering using readiness probes
- Recreate containers from the KV store in case of failure
- Templating of the container objects to leverage secrets and configuration
- Secrets, Configuration and Resources objects for external configuration


Installation
--------------------------

### (Bash) smrmgr
The smrmgr is bash script for management of the simplecontainer. It is used for:
- Downloading and installing client
- Starting the node in single or cluster mode
- Starting the node and joining to the existing cluster
- Various options and configuration simplified

```bash
curl -sL https://raw.githubusercontent.com/simplecontainer/smr/refs/heads/main/scripts/production/smrmgr.sh -o smrmgr
chmod +x smrmgr
sudo mv smrmgr /usr/local/bin
sudo smrmgr install
```

### (CLI) smr
The smr is client used to communicate to the local/external simplecontainer agents running on nodes.
The smrmgr automatically downloads the client and places it under `/usr/local/bin/smr`.

To manually install, start and manage simplecontainer nodes download the client from the releases:

https://github.com/simplecontainer/client/releases

Explore `/scripts/production/smrmgr.sh` to see how you can utilize smr client for the configuring and starting simplecontainer nodes.

### Running simplecontainer in cluster mode
Simplecontainer can run in single and cluster mode. Cluster mode allows users to deploy Docker daemons on different hosts and connect them via simplecontainer. An overlay network is created using flannel to enable inter-host communication.

Simplecontainer uses RAFT protocol to enable distributed state using the Badger key-value store.
Etcd embedded is also started in single mode and exposed to localhost only without credentials. This is because Flannel uses only Etcd as the state store for the network configuration.

Control-plane and RAFT communication is secured using mTLS so data is encrypted even over non-secure underlying networks.

Ports exposed:
- `0.0.0.0:1443->1443/tcp` (Simplecontainer control plane)
- `0.0.0.0:9212->9212/tcp` (RAFT protocol control plane sharing state)
- `:::1443->1443/tcp` (Simplecontainer control plane ipv6)
- `127.0.0.1:2379->2379/tcp` (Etcd exposed only on the localhost)

#### How to run it?

> [!IMPORTANT]
> The smrmgr script must be run on the host directly.

This scenario assumes there are two nodes(virtual machines) connected over non-secure internet connection.

- Node 1: node1.simplecontainer.com -> Points to Node 1 IP address
- Node 1: node2.simplecontainer.com -> Points to Node 2 IP address

**Node 1**
Requirements: 
- smgrmgr already installed.
- Docker daemon running on the Node 1

```bash
smrmgr start -a smr-agent-1 -d smr1.example.com -n https://node1.example.com
smrmgr export https://node1.example.com # Copy CONTEXT
cat $HOME/smr/smr/contexts/$(smr context).key # Copy KEY
```

**Node 2 (And any other nodes joining cluster)**
Requirements:
- smgrmgr already installed.
- Docker daemon running on the Node 1

```bash
smrmgr import {{ PASTE CONTEXT }} <<< {{ PASTE KEY }}
smrmgr import ca

smrmgr start -a smr-agent-2 -d smr2.example.com -n https://node2.example.com -j https://node1.example.com:1443
```

Afterward, cluster is started. Badger key-value store is now distributed using RAFT protocol. Flannel will start and agent will create docker network named `cluster`. Containers started are automatically connected to the flannel network when started.

### Running simplecontainer in single mode
[#### How to run it? (Exposed control plane to the internet/network using domain)
Exposing the control plane to the `0.0.0.0:1443` and `smr.example.com` will be only valid domain for the certificate authentication (**Change domain to your domain**):
]()
```bash
smrmgr start -a smr-agent-1 -d smr.example.com
smr context export smr-agent-1
```

From external machine run:
```bash
smr context import smr-agent-1 < {{ PASTE HERE EXPORTED CONTEXT OF smr-agent-1 }}
smr ps
```

#### How to run it? (Exposed control plane to the localhost only)
Exposing the control plane only to the localhost:

```bash
smrmgr start -a smr-agent-1 -e localhost:1443
```

#### How to run it? (Exposed control plane to the internet/network using IP)

Same as before smr client can be used:
```
smrmgr start -a smr-agent-1 -d smr.example.com -i 1.1.1.1
```


> [!IMPORTANT]
> Only domains and IPs listed in the smrmgr can verify against the simplecontainer agent.

Contexts and control plane
--------------------------

To access the simplecontainer control plane via local or public network, context needs to be added with the appropriate mtls bundle generated.

> [!IMPORTANT]
> Context for the localhost on the same machine is automatically imported if using smrmgr.

One can manually add the context to the client or another machine:
```bash
# Manually adding context
smr context connect https://localhost:1443 $HOME/.ssh/simplecontainer/root.pem --context smr-agent-1
{"level":"info","ts":1720694421.2032707,"caller":"context/Connect.go:40","msg":"authenticated against the smr-agent"}
```

Or import context which was exported from another agent to the client or another machine:
```bash
# Importing context
smr context import smr-agent-1 < {{ PASTE HERE EXPORTED CONTEXT OF smr-agent-1 }}
```

Afterward, smr can speak with nodes via control plane.

```bash
smr ps
GROUP  NAME  DOCKER NAME  IMAGE  IP  PORTS  DEPS  DOCKER STATE  SMR STATE
```

Access to the control plane of the simplecontainer is configured successfully if you get similar output.

## Running Docker containers using GitOps

It is possible to keep definition YAML files in the repository and let the simplecontainer apply it from the repository.

```bash
smr apply https://raw.githubusercontent.com/simplecontainer/examples/refs/heads/main/tests/minimal/definition.yaml
```

Applying this definition will create GitOps object on the simplecontainer.

```bash
smr gitops list
GROUP     NAME          REPOSITORY                                             REVISION  SYNCED        AUTO   STATE    
examples  plain-manual  https://github.com/simplecontainer/examples (cb849c3)  main      cb849c3       false  InSync  

smr gitops sync test smr

smr ps 
GROUP    NAME     DOCKER NAME        IMAGE           IP  PORTS  DEPS  DOCKER STATE  SMR STATE         
example  busybox  example-busybox-1  busybox:latest                   running       running (50m40s)  
example  busybox  example-busybox-2  busybox:latest                   running       running (50m40s)  
```

In this example, auto sync is disabled and needs to be triggered manually. When triggered the reconciler will apply 
all the definitions in the `/tests/minimal` directory from the `https://github.com/simplecontainer/examples` repository.

To see more info about the Gitops object:

```bash
smr gitops get examples plain-manual
```

Output:

```json
{
  "gitops": {
    "meta": {
      "group": "examples",
      "name": "plain-manual"
    },
    "spec": {
      "API": "",
      "automaticSync": false,
      "certKeyRef": {
        "Group": "",
        "Name": ""
      },
      "context": "",
      "directory": "/tests/minimal",
      "httpAuthRef": {
        "Group": "",
        "Name": ""
      },
      "poolingInterval": "",
      "repoURL": "https://github.com/simplecontainer/examples",
      "revision": "main"
    }
  },
  "kind": "gitops"
}
```

## Running containers (Plain way)

Run the next commands:
```bash
smr secret create secret.mysql.mysql.password 123456789
smr apply https://raw.githubusercontent.com/simplecontainer/examples/main/tests/simple-dependency-readiness/mysql-config.yaml
smr apply https://raw.githubusercontent.com/simplecontainer/examples/main/tests/simple-dependency-readiness/mysql-envs.yaml
smr apply https://raw.githubusercontent.com/simplecontainer/examples/main/tests/simple-dependency-readiness/nginx-config.yaml
smr apply https://raw.githubusercontent.com/simplecontainer/examples/main/tests/simple-dependency-readiness/traefik-config.yaml
smr apply https://raw.githubusercontent.com/simplecontainer/examples/main/tests/simple-dependency-readiness/containers.yaml
```

This example demonstrates:
- configuration
- resource
- container
- readiness check
- dependency

After running commands above, check the `smr ps`:
```bash
smr ps
GROUP    NAME     DOCKER NAME        IMAGE         IP                                      PORTS                      DEPS      DOCKER STATE  SMR STATE         
mysql    mysql    mysql-mysql-1      mysql:8.0     10.10.0.3 (ghost), 172.17.0.4 (bridge)  3306                                 running       running (51m17s)  
mysql    mysql    mysql-mysql-2      mysql:8.0     10.10.0.2 (ghost), 172.17.0.3 (bridge)  3306                                 running       running (51m15s)  
nginx    nginx    nginx-nginx-1      nginx:1.23.3  10.10.0.6 (ghost), 172.17.0.6 (bridge)  80, 443                    mysql.*   running       running (51m14s)  
traefik  traefik  traefik-traefik-1  traefik:v2.5  10.10.0.5 (ghost), 172.17.0.5 (bridge)  80:80, 443:443, 8888:8080  mysql.*   running       running (51m15s)  
```

Containers from group mysql will start first. 
Traefik and nginx will wait till mysql is ready because of the dependency definition and ordering.

Important links
---------------------------
- https://github.com/simplecosntainer/smr
- https://github.com/simplecontainer/client
- https://github.com/simplecontainer/examples
- https://simplecontainer.qdnqn.com

# License
This project is licensed under the GNU General Public License v3.0. See more in LICENSE file.