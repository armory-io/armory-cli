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
go build -ldflags "-X main.version=${VERSION}" -o ${DIST_DIR}/${APP_NAME} main.go
${DIST_DIR}/${APP_NAME} $@