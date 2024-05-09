Quick start
===========

**Note: The project is in active development.**

Last updated on  May 7, 2024

This is a quick start tutorial for getting a simple container manager up and running. Abbreviation is used as smr often.

Requirements
------------

*   Go version 1.22
*   Installed docker daemon on the environment
*   Sudo privileges

Installation of the agent
-------------------------

GitHub is used as the Git repository for the project. Currently, no release is made since the project is in active development.

To build smr just run these commands:

    git clone https://github.com/qdnqn/smr
    cd smr
    ./scripts/build_docker.sh
    ./scripts/run_docker_daemon.sh

The smr relies on the docker and it acts as a wrapper around the Docker API.

Running next command will show that smr-agent is up and running.

    docker ps
    CONTAINER ID   IMAGE       COMMAND                  CREATED         STATUS         PORTS                      NAMES
    e33a624da8d3   smr:0.0.1   "/bin/sh -c '/opt/sm…"   4 seconds ago   Up 4 seconds   127.0.0.1:8080->8080/tcp   smr-agent


The smr-agent is up and running and is listening on the local interface only so no remote connection is possible at the moment.

💡

The smr-agent is running as the docker container with super privileges to the /var/run/docker.sock to be able to manipulate docker daemon.

Installation of the client
--------------------------

The client also can be cloned from GitHub and needs to be built and copied to the directory that is already in the $PATH variable.

    git clone https://github.com/qdnqn/smr-client
    cd smr-client
    go build
    sudo cp smr /usr/local/bin/smr

Running containers with smr
---------------------------

To create a definition for the container it is kind of a mix of Kubernetes definition and Docker compose definition.

    kind: containers
    containers:
      traefik:
        meta:
          name: traefik
          group: traefik
        spec:
          options:
            enabled: false
          container:
            image: "traefik"
            tag: "v2.5"
            replicas: 1
            networks:
              - "demo"
            volumes:
              - host: "/var/run/docker.sock"
                target: "/var/run/docker.sock"
            ports:
              - container: "80"
                host: "80"
              - container: "443"
                host: "443"
              - container: "8080"
                host: "8888"

definition-traefik.yaml

After saving the file and running the next command:

    smr apply definition.yaml

💡

The smr-agent and smr-client must be on the same machine to communicate. Currently smr-agent is not exposed on the internet since it does not provide any auth method.

The agent will pick up the definition and it will create the container via Docker API.

Running `smr ps` or `docker ps` will return you the new state of the containers.

The Traefik container should be up and running.

    smr ps
    Group    Name     Image         IPs                    Ports                     Dependencies  Status              
    traefik  traefik  traefik:v2.5  172.17.0.3 10.10.0.2   80:80 443:443 8888:8080                 Dependency solved   


Or

    docker ps
    CONTAINER ID   IMAGE          COMMAND                  CREATED              STATUS          PORTS                                                              NAMES
    6962e21bdb6a   traefik:v2.5   "/entrypoint.sh trae…"   About a minute ago   Up 59 seconds   0.0.0.0:80->80/tcp, 0.0.0.0:443->443/tcp, 0.0.0.0:8888->8080/tcp   smr-traefik-traefik-1
    e33a624da8d3   smr:0.0.1      "/bin/sh -c '/opt/sm…"   4 hours ago          Up 4 hours      127.0.0.1:8080->8080/tcp                                           smr-agent

All the containers controlled via smr will have `smr` prefix and will follow the next naming notation: `smr-group-name-replicaNumber`.

Other containers can coexist with the one created using smr. Smr uses labels to track containers controlled via smr.

For example, if we run `docker stop smr-traefik-traefik-1` the container will be recreated and started again. The defined state will try to be applied. If the container fails five times consecutively it will be marked as BackOff - dead.

The definition of the containers follows an almost identical one from Docker compose.

Let's add some configuration.

    kind: resource
    resource:
      meta:
        group: traefik
        identifier: "*"
      spec:
        data:
          traefik-configuration: |
            providers:
              docker:
                exposedByDefault: false
            
            api:
              insecure: true
              dashboard: true


resource-traefik.yaml

The resource is kind of an object smr can understand. It exists as a standalone object something like ConfigMap or Secret.

To use it in the containers hit `smr apply resource-traefik.yaml` and redefine the container definition.

    kind: containers
    containers:
      traefik:
        meta:
          name: traefik
          group: traefik
        spec:
          options:
            enabled: false
          container:
            image: "traefik"
            tag: "v2.5"
            replicas: 1
            networks:
              - "demo"
            volumes:
              - host: "/var/run/docker.sock"
                target: "/var/run/docker.sock"
            ports:
              - container: "80"
                host: "80"
              - container: "443"
                host: "443"
              - container: "8080"
                host: "8888"
            resources:
              - identifier: "*"
                key: traefik-configuration
                mountPoint: /etc/traefik/traefik.yml

definition-traefik.yaml

Hitting `smr apply definition-traefik.yaml` will recreate the Traefik container and will mount the resource inside the container at the `/etc/traefik/traefik.yml`.

Another object available for use is Configuration.

    kind: configuration
    configuration:
      meta:
        group: mysql
        identifier: "*"
      spec:
        data:
          password: "password123"

Configuration object is also a standalone object and can be used to define secret or configuration variables for the container which later can be used at the runtime.

    kind: containers
    containers:
      traefik:
        meta:
          name: traefik
          group: traefik
        spec:
          options:
            enabled: false
          container:
            image: "traefik"
            tag: "v2.5"
            replicas: 1
            networks:
              - "demo"
            volumes:
              - host: "/var/run/docker.sock"
                target: "/var/run/docker.sock"
            ports:
              - container: "80"
                host: "80"
              - container: "443"
                host: "443"
              - container: "8080"
                host: "8888"
            resources:
              - identifier: "*"
                key: traefik-configuration
                mountPoint: /etc/traefik/traefik.yml
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
              - "demo"
            configuration:
              password: "{{ configuration.mysql[*].password }}"

definition-traefik.yaml

Hitting `smr apply definition-traefik.yaml` will now only create mysql container since Traefik's definition is the same.

As you can see container definition is templatable using `{{ }}` notation to access the runtime data or another object data.

Features included in the smr:

*   Own DNS resolving - [Read more here](https://smr.qdnqn.com/dns/)
*   Gitops - [Read more here](https://smr.qdnqn.com/gitops/)
*   Key-Value store holding all data - [Read more here](https://smr.qdnqn.com/key-value-store/)
*   Replicas and headless service - [Read more here](https://smr.qdnqn.com/replicas/)
*   [Configuration](https://smr.qdnqn.com/configuration/) and [resources](https://smr.qdnqn.com/resource/)
*   Templating of the container objects with runtime information - [Read more here](https://smr.qdnqn.com/templating-container-object/)
*   Dependencies between containers - [Read more here](https://smr.qdnqn.com/container-dependencies/)
*   Operators - [Read more here](https://smr.qdnqn.com/operators/)
*   Reconciliation - [Read more here](https://smr.qdnqn.com/reconciliation/)
*   [Smr agent](https://smr.qdnqn.com/reconciliation/) and [smr client](https://github.com/qdnqn/smr-client?ref=smr.qdnqn.com)

You can also check official repositories.

https://github.com/qdnqn/smr?ref=smr.qdnqn.com)
https://github.com/qdnqn/smr-client?ref=smr.qdnqn.com)

Or if you want more examples hit the examples repository.

https://github.com/qdnqn/smr-examples?ref=smr.qdnqn.com

# License
This project is licensed under the GNU General Public License v3.0. See more in LICENSE file.