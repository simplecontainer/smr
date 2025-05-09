build-engine:
	./development/build-engine.sh

build-ctl:
	./development/build-ctl.sh

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

install:
	./development/install-ctl.sh
	./development/install-engine.sh

install-ctl:
	./development/install-ctl.sh

install-engine:
	./development/install-engine.sh

docker: build-docker
standalone: build-engine build-ctl install build-docker run-standalone
ctl: build-ctl install-ctl
cluster: build-engine  build-ctl install build-docker run-cluster
cluster-debug: build-docker-debug run-cluster-debug