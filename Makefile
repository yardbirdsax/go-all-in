WEB_IMAGE_NAME=golang-web
TEST_IMAGE_NAME=golang-web-test
DOCKER_NETWORK_NAME=golang-web
VERSION=1
GO_IMAGE_NAME=golang
GO_VERSION=1.18
GOBIN_PATH = $$PWD/.bin
ENV_VARS = GOBIN="$(GOBIN_PATH)" PATH="$(GOBIN_PATH):$$PATH"
.SILENT:

export SHELL:=/bin/bash
export SHELLOPTS:=$(if $(SHELLOPTS),$(SHELLOPTS):)pipefail:errexit

.ONESHELL:

.PHONY: docker-build-api
docker-build-api:
	docker build -t $(WEB_IMAGE_NAME) --build-arg VERSION=$(VERSION) -f ./Dockerfile.api .
.PHONY: docker-build-test
docker-build-test:
	docker build -t $(TEST_IMAGE_NAME) --build-arg VERSION=$(VERSION) -f ./Dockerfile.api.integration .
.PHONY: docker-build
docker-build: docker-build-web docker-build-test
.PHONY: docker-create-network
docker-create-network:
	if [ -z "$$(docker network ls | grep $(DOCKER_NETWORK_NAME))" ]; then \
		docker network create $(DOCKER_NETWORK_NAME); \
	fi
.PHONY: docker-rm-network
docker-rm-network:
	if [ -n "$$(docker network ls | grep $(DOCKER_NETWORK_NAME))" ]; then \
		docker network rm $(DOCKER_NETWORK_NAME); \
	fi
.PHONY: docker-start-web
docker-start-web: docker-create-network docker-build-api
	docker run --rm -d --name $(WEB_IMAGE_NAME) --network $(DOCKER_NETWORK_NAME) $(WEB_IMAGE_NAME)
.PHONY: docker-start-test
docker-start-test: docker-create-network docker-build-test
	docker run \
		--rm \
		-e WEB_URL=http://$(WEB_IMAGE_NAME):8080 \
		-e WEB_EXPECTED_VERSION="$(VERSION)" \
		--name $(TEST_IMAGE_NAME) \
		--network $(DOCKER_NETWORK_NAME) \
		$(TEST_IMAGE_NAME)
.PHONY: docker-stop-web
docker-stop-web:
	if [ -n "$$(docker ps | grep $(WEB_IMAGE_NAME))" ]; then \
		docker stop $(WEB_IMAGE_NAME); \
	fi

.PHONY: generate
generate:
	$(ENV_VARS) go generate ./...

.PHONY: run-integration-test
run-integration-test:
	function tearDown {
		$(MAKE) -k docker-stop-web
		$(MAKE) -k docker-rm-network
	}
	trap tearDown EXIT

	$(MAKE) docker-create-network
	$(MAKE) docker-start-web
	$(MAKE) docker-start-test

.PHONY: test
test:
	go test -v -count=1 -cover -tags unit -coverprofile coverage.out ./...

.PHONY: tool
tools:
	$(ENV_VARS) go install $$(go list -f '{{join .Imports " "}}' tools.go)