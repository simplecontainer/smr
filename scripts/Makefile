build:
	./development/build.sh

build-docker:
	./development/docker-build.sh

build-docker-debug:
	./development/docker-build-debug.sh

run-standalone:
	./development/docker-run-standalone.sh

run-cluster:
	./development/docker-run-cluster.sh

run-cluster-debug:
	./development/docker-run-cluster-debug.sh

build: build
docker: build-docker
standalone: build build-docker run-standalone
cluster: build build-docker run-cluster
cluster-debug: build-docker-debug run-cluster-debug