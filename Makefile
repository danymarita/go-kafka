.PHONY: default help build package tag push helm helm-migration run test clean

SHELL         = /bin/bash
APP_NAME      = go_kafka
VERSION      := $(shell git describe --always --tags)
GIT_COMMIT    = $(shell git rev-parse HEAD)
GIT_DIRTY     = $(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)
BUILD_DATE    = $(shell date '+%Y-%m-%d-%H:%M:%S')

default: help

help:
	@echo 'Management commands for ${APP_NAME}:'
	@echo
	@echo 'Usage:'
	@echo '    make build_api                 Compile API for writer.'
	@echo '    make build_worker              Compile worker for reader.'
	@echo '    make run_api                   Run API for writer.'
	@echo '    make run_worker                Run worker for reader.'

	@echo

build_api:
	@echo "Building ${APP_NAME}_api ${VERSION}"
	cd cmd/api && go build -ldflags "-w -X github.com/danymarita/go-kafka/version.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X github.com/danymarita/go-kafka/version.Version=${VERSION} -X github.com/danymarita/go-kafka/version.Environment=${ENVIRONMENT} -X github.com/danymarita/go-kafka/version.BuildDate=${BUILD_DATE}" -o ../../${APP_NAME}_api

build_worker:
	@echo "Building ${APP_NAME}_worker ${VERSION}"
	cd cmd/worker && go build -ldflags "-w -X github.com/danymarita/go-kafka/version.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X github.com/danymarita/go-kafka/version.Version=${VERSION} -X github.com/danymarita/go-kafka/version.Environment=${ENVIRONMENT} -X github.com/danymarita/go-kafka/version.BuildDate=${BUILD_DATE}" -o ../../${APP_NAME}_worker

run_api: build_api
	@echo "Running ${APP_NAME}_api ${VERSION}"
	./${APP_NAME}_api

run_worker: build_worker
	@echo "Running ${APP_NAME}_api ${VERSION}"
	./${APP_NAME}_worker