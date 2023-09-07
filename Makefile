APP_NAME              = armory
APP_EXT               ?= "${CLI_EXT}"
VERSION               ?= $(shell ./scripts/version.sh)
REGISTRY              ?="armory-docker-local.jfrog.io"
REGISTRY_ORG          ?="armory"
GOARCH                ?= $(shell go env GOARCH)
GOOS                  ?= $(shell go env GOOS)
PWD                   =  $(shell pwd)
IMAGE_TAG             ?= $(VERSION)
LOCAL_KUBECTL_CONTEXT ?= "kind-armory-cloud-dev"
IMAGE                 := $(subst $\",,$(REGISTRY)/$(REGISTRY_ORG)/${APP_NAME}:${VERSION})
BUILD_DIR             := ${PWD}/build
DIST_DIR              := ${BUILD_DIR}/dist/$(GOOS)_$(GOARCH)
GEN_DIR               := ${PWD}/generated
MAIN_PATH			  := "main.go"
TIMESTAMP			  := $(shell date -u +"%FT%TZ")

default: all

include ./scripts/common_targets.mk

.PHONY: all
all: clean build-dirs run-before-tools build check run-after-tools

.PHONY: tools
tools:
	echo installing tools.... && \
	go install github.com/vakenbolt/go-test-report@v0.9.3 && \
	go install github.com/undoio/delve/cmd/dlv@latest && \
	echo installing static check... && \
	go install honnef.co/go/tools/cmd/staticcheck@latest


.PHONY: configure-build
configure-build:
	@go env -w CGO_ENABLED=0
	@go env

.PHONY: integration
integration: build-dirs install-tools
	@go test -v -cover ./integration/... -json > ${BUILD_DIR}/reports/integration-test-report.json
	@go test -v -coverprofile=${BUILD_DIR}/reports/integration.cov ./integration/...
	@cat ${BUILD_DIR}/reports/integration-test-report.json | go-test-report --title ${APP_NAME}-integration-test -v --output ${BUILD_DIR}/reports/integration_test_report.html

.PHONY: release
release: clean build-linux-amd64
	@echo Release version of armory-cli ${VERSION} created in ${DIST_DIR}/${APP_NAME}${APP_EXT}
	@echo $(TIMESTAMP) $(IMAGE_TAG)
	@docker build --tag $(REGISTRY_ORG)/${APP_NAME}-cli:$(IMAGE_TAG) \
	--label "org.opencontainers.image.created=$(TIMESTAMP)" \
	--label "org.opencontainers.image.description=The CLI for Armory Continuous Deployments-as-a-Service" \
	--label "org.opencontainers.image.revision=$(GITHUB_SHA)" \
	--label "org.opencontainers.image.licenses=Apache-2.0" \
	--label "org.opencontainers.image.source=https://github.com/armory-io/armory-cli" \
	--label "org.opencontainers.image.title=armory-cli" \
	--label "org.opencontainers.image.url=https://github.com/armory-io/armory-cli" \
	--label "org.opencontainers.image.version=$(VERSION)" \
	-f Dockerfile . \
	--push


