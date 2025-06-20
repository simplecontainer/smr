.PHONY: test clean

HOME ?=
IMAGE ?= smr
TAG ?= latest
BINARY ?=
BINARY_CLI ?=
TIMEOUT ?= 60s
CLEANUP ?= true
TEST_NAME ?= # Empty by default, will run all tests if not specified
TEST_DIR ?= ./tests/integration/... # Will run all integration tests if not specified
EXAMPLES_DIR ?=

# Common test flags
TEST_FLAGS = -image=$(IMAGE) -tag=$(TAG) -timeout=$(TIMEOUT)
ifneq ($(BINARY),)
TEST_FLAGS += -binary=$(BINARY)
endif

ifneq ($(BINARY_CLI),)
TEST_FLAGS += -binaryctl=$(BINARY_CLI)
endif

ifneq ($(HOME),)
TEST_FLAGS += -root=$(HOME)
endif

ifneq ($(EXAMPLES_DIR),)
TEST_FLAGS += -examples=$(EXAMPLES_DIR)
endif

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

nuke:
	./development/nuke.sh

test-e2e:
	@echo "Running test(s): $(if $(TEST_NAME),$(TEST_NAME),all tests)"
ifneq ($(TEST_NAME),)
	go test -v -p 1 -tags=e2e ../pkg/tests/e2e/$(TEST_DIR) -run $(TEST_NAME) $(TEST_FLAGS)
else
	go test -v -p 1 -tags=e2e ../pkg/tests/e2e/$(TEST_DIR) $(TEST_FLAGS)
endif

ctl: build-ctl install-ctl
docker: nuke build-engine build-ctl install build-docker
standalone: nuke build-engine build-ctl install build-docker run-standalone
cluster: nuke build-engine build-ctl install build-docker run-cluster
cluster-debug: nuke build-docker-debug run-cluster-debug