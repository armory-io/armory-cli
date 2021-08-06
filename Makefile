VERSION_TYPE    ?= "snapshot" # Must be one of: "snapshot", "rc", or "release"
VERSION  ?= $(shell build/version.sh)
REGISTRY ?= "docker.io"
REGISTRY_ORG ?= "armory"
GOARCH    ?= $(shell go env GOARCH)
GOOS      ?= $(shell go env GOOS)
UNAME_S := $(shell uname -s)
NAMESPACE ?= "default"
BUF_VERSION = 0.41.0
PWD = $(shell pwd)

PKG             := github.com/armory/armory-cli
SRC_DIRS        := cmd internal
BUILD_DIR       := ${PWD}/dist/$(GOOS)_$(GOARCH)

.PHONY: all
all: build

############
## Building
############
.PHONY: build-dirs
build-dirs:
	@mkdir -p $(BUILD_DIR)


.PHONY: build
build: build-dirs Makefile
	@echo "Building ${BUILD_DIR}/armory${CLI_EXT}..."
	@go build ${LDFLAGS} -o ${BUILD_DIR}/armory${CLI_EXT} main.go

############
## Testing
############
.PHONY: test
test: build-dirs Makefile
	@go test -cover ./...

.PHONY: coverage
coverage:
	@go test -coverprofile=profile.cov ./...
	@go tool cover -html=profile.cov

.PHONY: version
version:
	@echo $(VERSION)

.PHONY: lint
lint:
	@find pkg cmd -name '*.go' | grep -v 'generated' | xargs -L 1 golint

.PHONY: clean
clean:
	rm -rf dist