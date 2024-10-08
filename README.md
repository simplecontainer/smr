![simplecontainer manager](.github/resources/repository.jpg)

Quick start
===========

> [!WARNING]
> The project is not stable yet. Releases and major changes are introduced often. 

This is a quick start tutorial for getting a simple container up and running.

## Description
A simple container manager is designed to ease life for the developers and DevOps engineers running containers on Docker.

Introducing objects which can be defined as YAML definition and sent to the simplecontainer manager to produce Docker container via reconciliation:

- Containers
- Container
- Configuration
- Resource
- Gitops
- CertKey
- HttpAuth

These objects let you manage Docker containers with configure features:

- Single Docker daemon only (Currently)
- Integrated DNS server isolated from Docker daemon
- GitOps: deploy objects from the GitOps repositories
- Replication of containers
- Reconciliation and tracking the lifecycle of the Docker containers
- Operators to implement third-party functionalities
- CLI client to interact with the simplecontainer manager
- Fast learning curve - no over complication
- Reliable dependency ordering and readiness probes
- Recreate containers from the KV store in case of failure
- Templating of the container objects to leverage secrets and configuration


Installation of the agent
-------------------------
To start using simple container first run it to generate smr project and build configuration file.

```bash
LATEST_VERSION=$(curl -s https://raw.githubusercontent.com/simplecontainer/smr/main/version)

mkdir $HOME/.smr
docker pull simplecontainermanager/smr:$LATEST_VERSION
docker run \
       -v $HOME/.smr:/home/smr-agent/smr \
       -e DOMAIN=localhost \
       -e EXTERNALIP=127.0.0.1 \
       simplecontainermanager/smr:$LATEST_VERSION create smr
```

This will generate project and create configuration file.

Afterward running will start simplecontainer as docker container, and it will be able
to manage containers on top of docker.

```bash
LATEST_VERSION=$(curl -s https://raw.githubusercontent.com/simplecontainer/smr/main/version)

docker run \
       -v /var/run/docker.sock:/var/run/docker.sock \
       -v $HOME/.smr:/home/smr-agent/smr \
       -v $HOME/.ssh:/home/smr-agent/.ssh \
       -v /tmp:/tmp \
       -p 0.0.0.0:1443:1443 \
       --dns 127.0.0.1 \
       --name smr-agent \
       -d simplecontainermanager/smr:$LATEST_VERSION start
```

Agent will generate certificates under `$HOME/.ssh/simplecontainer`. These are important and used by the client to communicate
with the simplecontainer agent in a secured manner. These certificates are only valid for the root user and this certificate bundle
is only usable from the localhost to be more secured.

This bundle is needed by the client to connect to the Simplecontainer API.

Adding new certificates can be done via CLI for users to connect from external domains.
```bash
cat $HOME/.ssh/simplecontainer/client.pem
-----BEGIN PRIVATE KEY-----
MIIJQwIBADANBgkqhkiG9w0BAQEFAASCCS0wggkpAgEAAoICAQDBNozIEBzUyvJf
ln8CH/I1cX6W/EzX+SNh/WYD2pYiCkgKgRUdPNrua7Vf3/zPrNmAqdHyQgDIjNlr
...
```

Installation of the client
--------------------------

Client CLI is used for communication to the simplecontainer over network using mTLS. 
It is secured by mutual verification and encryption.

To install client just download it from releases:

https://github.com/simplecontainer/client/releases

Example for installing latest version:

```bash
export VERSION=$(curl -s https://raw.githubusercontent.com/simplecontainer/client/main/version)
export PLATFORM=linux-amd64
curl -Lo client https://github.com/simplecontainer/client/releases/download/$VERSION/client-$PLATFORM
chmod +x client
sudo mv client /usr/local/bin/smr
smr context connect https://localhost:1443 $HOME/.ssh/simplecontainer/client.pem --context localhost
{"level":"info","ts":1720694421.2032707,"caller":"context/Connect.go:40","msg":"authenticated against the smr-agent"}
smr ps
GROUP  NAME  DOCKER NAME  IMAGE  IP  PORTS  DEPS  DOCKER STATE  SMR STATE
```
Afterward access to control plane of the simple container is configured.

## Running Docker containers using GitOps

It is possible to keep definition YAML files in the repository and let the simplecontainer apply it from the repository.

```bash
smr apply https://raw.githubusercontent.com/simplecontainer/examples/main/gitops/gitops-plain.yaml 
```

Applying this definition will create GitOps object on the simplecontainer.

```bash
smr gitops list                               
GROUP  NAME     REPOSITORY                                   REVISION  SYNCED        AUTO   STATE    
test   smr      https://github.com/simplecontainer/examples  main      Never synced  false  Drifted  

smr gitops sync test smr

smr ps 
GROUP    NAME     DOCKER NAME        IMAGE         IP                                      PORTS                      DEPS  DOCKER STATE  SMR STATE  
nginx    nginx    nginx-nginx-1      nginx:1.23.3  10.10.0.3 (ghost), 172.17.0.3 (bridge)  80, 443                          running        (2m0s)    
nginx    nginx    nginx-nginx-2      nginx:1.23.3  10.10.0.4 (ghost), 172.17.0.4 (bridge)  80, 443                          running        (2m0s)    
nginx    nginx    nginx-nginx-3      nginx:1.23.3  10.10.0.5 (ghost), 172.17.0.5 (bridge)  80, 443                          running        (2m0s)    
traefik  traefik  traefik-traefik-1  traefik:v2.5  10.10.0.6 (ghost), 172.17.0.6 (bridge)  80:80, 443:443, 8888:8080        running        (2m0s)    
```

In this example, auto sync is disabled and needs to be triggered manually. When triggered the reconciler will apply 
all the definitions in the `/gitops/bundle` directory from the `https://github.com/simplecontainer/examples` repository.

To see more info about the Gitops object:

```bash
smr gitops get test smr
```

Output:

```json
{
  "gitops": {
    "meta": {
      "group": "test",
      "name": "smr"
    },
    "spec": {
      "automaticSync": false,
      "certKeyRef": {
        "Group": "",
        "Identifier": ""
      },
      "directory": "/gitops/bundle",
      "httpAuthRef": {
        "Group": "",
        "Identifier": ""
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
Traefik and nginx will wait till mysql is ready because of the dependency defined.

Important links
---------------------------
- https://github.com/simplecontainer/smr
- https://github.com/simplecontainer/client
- https://github.com/simplecontainer/examples
- https://smr.qdnqn.com

# License
This project is licensed under the GNU General Public License v3.0. See more in LICENSE file.