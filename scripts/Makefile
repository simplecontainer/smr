build:
	./development/build.sh

build_production:
	./development/build_production.sh

build_docker:
	./development/docker_build.sh

run:
	./development/docker_run.sh

run_private:
	./development/docker_run_private.sh

develop: build build_docker run
production: build_production build_docker run
dockerd: build_docker run
dockerp: build_docker run_private
