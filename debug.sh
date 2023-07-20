#!/usr/bin/env bash

set -e

APP_NAME=armory
GOARCH=$(go env GOARCH)
GOOS=$(go env GOOS)
BUILD_DIR=${PWD}/build
DIST_DIR=${BUILD_DIR}/bin/${GOOS}_${GOARCH}
VERSION=dev

export ADDITIONAL_ACTIVE_PROFILES="local-overrides"
export APPLICATION_NAME=${APP_NAME}
export APPLICATION_VERSION=${VERSION}
go build -ldflags "-X main.version=${VERSION}" -o ${DIST_DIR}/${APP_NAME}-debug -gcflags "all=-N -l" main.go
dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec ${DIST_DIR}/${APP_NAME}-debug $@