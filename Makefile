GOARCH    ?= $(shell go env GOARCH)
GOOS      ?= $(shell go env GOOS)
PWD = $(shell pwd)


BUILD_DIR       := ${PWD}/build
DIST_DIR        := ${BUILD_DIR}/dist/$(GOOS)_$(GOARCH)
REPORTS_DIR     := ${BUILD_DIR}/reports/coverage

.PHONY: all
all: pre version clean test coverage build

############
## Building
############
.PHONY: build-dirs
build-dirs:
	@mkdir -p ${BUILD_DIR}
	@mkdir -p ${DIST_DIR}
	@mkdir -p ${REPORTS_DIR}

.PHONY: pre
pre:
#	@go env -w GOPRIVATE=github.com/armory-io/deploy-engine
	@git config --global http.extraheader "authorization: Bearer {GITHUB_TOKEN}"
	@go env
	@GOPRIVATE=github.com/armory-io/deploy-engine go get github.com/armory-io/deploy-engine@v0.2.0

.PHONY: build
build: build-dirs Makefile
	@echo "Building ${DIST_DIR}/armory${CLI_EXT}..."
	@go build -v -ldflags="-X 'github.com/armory/armory-cli/cmd/version.Version=${VERSION}' -X 'github.com/armory/armory-cli/cmd/login.UserClientId=${USER_CLIENT_ID}'" -o ${DIST_DIR}/armory${CLI_EXT} main.go

############
## Testing
############
.PHONY: test
test: build-dirs Makefile
	@go test -v -cover ./pkg/... ./cmd/... -json > test-report.json
.PHONY: coverage
coverage:
	@go test -v -coverprofile=profile.cov ./pkg/... ./cmd/...
	@go tool cover -html=profile.cov -o ${BUILD_DIR}/reports/coverage/index.html

.PHONY: version
version:
	@echo $(VERSION)

.PHONY: clean
clean:
	rm -rf dist

.PHONY: integration
integration: build-dirs Makefile
	@go test -v -cover ./integration/... -json > integration-test-report.json
	@go test -v -coverprofile=integration.cov ./integration/...