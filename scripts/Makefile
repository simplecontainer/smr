build-engine:
	./development/build_engine.sh

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

docker: build-docker
standalone: build-engine build-docker run-standalone
cluster: build-engine build-docker run-cluster
cluster-debug: build-docker-debug run-cluster-debug