build:
	./development/build.sh

build_production:
	./development/build_production.sh

build_docker:
	./development/docker_build.sh

run:
	./development/docker_run.sh

run_cluster:
	./development/docker_run_cluster.sh

run_private:
	./development/docker_run_private.sh

logs:
	docker logs -f smr-agent

develop: build build_docker run logs
cluster: build build_docker run_cluster
production: build_production build_docker run
dockerd: build_docker run
dockerp: build_docker run_private
run: run