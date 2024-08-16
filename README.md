Quick start
===========

**Note: The project is not stable.**

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
- Templating of the container objects to leverage secrets and configuration


Installation of the agent
-------------------------
To start using simple container first run it to generate smr project and build configuration file.

Note: This is example for the localhost. If domain is example.com running on the virtual machine with IP 1.2.3.4,
just replace the DOMAIN and EXTERNALIP values.

```bash
LATEST_VERSION=v0.0.1

mkdir $HOME/.smr
docker pull simplecontainermanager/smr:$LATEST_VERSION
docker run \
       -v $HOME/.smr:/home/smr-agent/smr \
       -e DOMAIN=localhost \
       -e EXTERNALIP=127.0.0.1 \
       smr:$LATEST_VERSION create smr
```

This will generate project and create configuration file.

Afterward running will start simplecontainer as docker container, and it will be able
to manage containers on top of docker.

```bash
LATEST_VERSION=v0.0.1

docker run \
       -v /var/run/docker.sock:/var/run/docker.sock \
       -v $HOME/.smr:/home/smr-agent/smr \
       -v $HOME/.ssh:/home/smr-agent/.ssh \
       -v /tmp:/tmp \
       -p 0.0.0.0:1443:1443 \
       --dns 127.0.0.1 \
       --name smr-agent \
       -d smr:$LATEST_VERSION start
```

This will generate certificates under `$HOME/.ssh/simplecontainer`. These are important and used by the client to communicate 
with the simplecontainer agent in a secured manner.

Installation of the client
--------------------------

Client CLI is used for communication to the simplecontainer over network using mTLS. 
It is secured by mutual verification and encryption.

To install client just download it from releases:

https://github.com/simplecontainer/client/releases

Example:

```azure
VERSION=v0.0.1
PLATFORM=linux-amd64
curl -o client https://github.com/simplecontainer/client/releases/download/$VERSION/client-$PLATFORM
sudo mv client /usr/bin/smr
smr context connect https://localhost:1443 $HOME/.ssh/simplecontainer/client.pem --context localhost
{"level":"info","ts":1720694421.2032707,"caller":"context/Connect.go:40","msg":"authenticated against the smr-agent"}
smr ps
GROUP  NAME  DOCKER NAME  IMAGE  IP  PORTS  DEPS  DOCKER STATE  SMR STATE
```
Afterward access to control plane of the simple container is configured.

## Running containers (Plain way)
Define containers.yaml file:
```yaml
kind: containers
meta:
  name: application-bundle
  group: app
spec:
  mysql:
    meta:
      name: mysql
      group: mysql
    spec:
      options:
        enabled: false
      container:
        image: "mysql"
        tag: "8.0"
        replicas: 1
        envs:
          - "MYSQL_ROOT_PASSWORD={{ configuration.password }}"
        networks:
          - "ghost"
        ports:
          - container: "3306"
        readiness:
          - name: "mysql.*"
            operator: DatabaseReady
            timeout: "20s"
            body:
              ip: "mysql.mysql-mysql-1.cluster.private"
              username: "{{ configuration.username }}"
              password: "{{ configuration.password }}"
              port: "3306"
        configuration:
          username: "root"
          password: "{{ configuration.mysql.*.password }}"
```
Define configuration-mysql.yaml file:
```yaml
kind: configuration
spec:
  meta:
    group: mysql
    identifier: "*"
  spec:
    data:
      password: "{{ secret.mysql.mysql.password }}"
```
Run the next commands:
```bash
smr secret create secret.mysql.mysql.password 123456789
smr apply configuration-mysql.yaml
smr apply containers.yaml
smr ps

```

## Running containers (GitOps way)

It is possible to hold definition YAML files in the repository and let the simplecontainer apply it from the repository.

```yaml
kind: gitops
spec:
  meta:
    group: test
    identifier: testApp
  spec:
    repoURL: "https://github.com/simplecontainer/examples"
    revision: "main"
    directoryPath: "/gitops/bundle"
```

Applying this definition will create GitOps object on the simplecontainer.

```bash
smr gitops list                               
GROUP  NAME     REPOSITORY                                   REVISION  SYNCED        AUTO   STATE    
test   testApp  https://github.com/simplecontainer/examples  main      Never synced  false  Drifted  

smr gitops test testApp sync 
```

In this example auto sync is disabled and needs to be triggered manually. When triggered the reconciler will apply 
all the definitions in the `/gitops/bundle` directory from the `https://github.com/simplecontainer/examples` repository.

Important links
---------------------------
- https://github.com/simplecontainer/smr
- https://github.com/simplecontainer/client
- https://github.com/simplecontainer/examples
- https://smr.qdnqn.com

# License
This project is licensed under the GNU General Public License v3.0. See more in LICENSE file.