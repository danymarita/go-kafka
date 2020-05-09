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
	@echo '    make build_api                 	Compile API for writer.'
	@echo '    make build_order_worker        Compile worker for order processor.'
	@echo '    make build_email_sender_worker   Compile worker for send email.'
	@echo '    make run_api                   	Run API for writer.'
	@echo '    make run_order_worker          Run worker for order processor.'
	@echo '    make run_email_sender_worker     Run worker for send email.'

	@echo

build_api:
	@echo "Building ${APP_NAME}_api ${VERSION}"
	cd cmd/api && go build -ldflags "-w -X github.com/danymarita/go-kafka/version.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X github.com/danymarita/go-kafka/version.Version=${VERSION} -X github.com/danymarita/go-kafka/version.Environment=${ENVIRONMENT} -X github.com/danymarita/go-kafka/version.BuildDate=${BUILD_DATE}" -o ../../${APP_NAME}_api

build_order_worker:
	@echo "Building ${APP_NAME}_order_worker ${VERSION}"
	cd cmd/worker/order && go build -ldflags "-w -X github.com/danymarita/go-kafka/version.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X github.com/danymarita/go-kafka/version.Version=${VERSION} -X github.com/danymarita/go-kafka/version.Environment=${ENVIRONMENT} -X github.com/danymarita/go-kafka/version.BuildDate=${BUILD_DATE}" -o ../../../${APP_NAME}_order_worker

build_email_sender_worker:
	@echo "Building ${APP_NAME}_email_sender_worker ${VERSION}"
	cd cmd/worker/email_sender && go build -ldflags "-w -X github.com/danymarita/go-kafka/version.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X github.com/danymarita/go-kafka/version.Version=${VERSION} -X github.com/danymarita/go-kafka/version.Environment=${ENVIRONMENT} -X github.com/danymarita/go-kafka/version.BuildDate=${BUILD_DATE}" -o ../../../${APP_NAME}_email_sender_worker

run_api: build_api
	@echo "Running ${APP_NAME}_api ${VERSION}"
	./${APP_NAME}_api

run_order_worker: build_order_worker
	@echo "Running ${APP_NAME}_order_worker ${VERSION}"
	./${APP_NAME}_order_worker

run_email_sender_worker: build_email_sender_worker
	@echo "Running ${APP_NAME}_email_sender_worker ${VERSION}"
	./${APP_NAME}_email_sender_worker