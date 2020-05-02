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
	@echo '    make build                 Compile the project.'
	@echo '    make package               Build final Docker image with just the Go binary inside.'
	@echo '    make tag                   Tag image created by package with latest, git commit and version.'
	@echo '    make push                  Push tagged images to registry.'
	@echo '    make helm                  Deploy to Kubernetes via Helm.'
	@echo '    make helm-migration        Run database migration via Helm.'
	@echo '    make run ARGS=             Run with supplied arguments.'
	@echo '    make test                  Run tests on a compiled project.'
	@echo '    make test-cover            Run tests with goveralls.'
	@echo '    make clean                 Clean the directory tree.'

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