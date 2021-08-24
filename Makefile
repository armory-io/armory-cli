GOARCH    ?= $(shell go env GOARCH)
GOOS      ?= $(shell go env GOOS)
PWD = $(shell pwd)

BUILD_DIR       := ${PWD}/build
DIST_DIR        := ${BUILD_DIR}/dist/$(GOOS)_$(GOARCH)
REPORTS_DIR     := ${BUILD_DIR}/reports/coverage

.PHONY: all
all: version clean test coverage build

############
## Building
############
.PHONY: build-dirs
build-dirs:
	@mkdir -p ${BUILD_DIR}
	@mkdir -p ${DIST_DIR}
	@mkdir -p ${REPORTS_DIR}


.PHONY: build
build: build-dirs Makefile
	@echo "Building ${DIST_DIR}/armory${CLI_EXT}..."
	@go build -ldflags="-X 'github.com/armory/armory-cli/cmd/version.Version=${VERSION}'" -o ${DIST_DIR}/armory${CLI_EXT} main.go

############
## Testing
############
.PHONY: test
test: build-dirs Makefile
	@go test -cover ./...

.PHONY: coverage
coverage:
	@go test -coverprofile=profile.cov ./...
	@go tool cover -html=profile.cov -o ${BUILD_DIR}/reports/coverage/index.html

.PHONY: version
version:
	@echo $(VERSION)

.PHONY: clean
clean:
	rm -rf dist