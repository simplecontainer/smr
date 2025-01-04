build:
	./development/build.sh

build_docker:
	./development/docker_build.sh

run_standalone:
	./development/docker_run_standalone.sh

run_cluster:
	./development/docker_run_cluster.sh

build: build
docker: build_docker
standalone: build build_docker run_standalone
cluster: build build_docker run_cluster