# SPDX-FileCopyrightText: 2023 Christoph Mewes
# SPDX-License-Identifier: MIT

GIT_VERSION = $(shell git describe --tags --always)
GIT_HEAD ?= $(shell git log -1 --format=%H)
NOW_GO_RFC339 = $(shell date --utc +'%Y-%m-%dT%H:%M:%SZ')

export CGO_ENABLED ?= 0
export GOFLAGS ?= -mod=readonly -trimpath
OUTPUT_DIR ?= _build
GO_DEFINES ?= -X main.BuildTag=$(GIT_VERSION) -X main.BuildCommit=$(GIT_HEAD) -X main.BuildDate=$(NOW_GO_RFC339)
GO_LDFLAGS += -w -extldflags '-static' $(GO_DEFINES)
GO_BUILD_FLAGS ?= -v -ldflags '$(GO_LDFLAGS)'
GO_TEST_FLAGS ?= -v -race

default: build

.PHONY: build
build:
	go build $(GO_BUILD_FLAGS) -o $(OUTPUT_DIR)/ ./cmd/dumper
	go build $(GO_BUILD_FLAGS) -o $(OUTPUT_DIR)/ ./cmd/swaggerdumper
	go build $(GO_BUILD_FLAGS) -o $(OUTPUT_DIR)/ ./cmd/render

.PHONY: test
test:
	CGO_ENABLED=1 go test $(GO_TEST_FLAGS) ./...

.PHONY: clean
clean:
	rm -rf $(OUTPUT_DIR)

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: full-rebuild
full-rebuild: make clean build dump-all

.PHONY: dump
dump: clean build
	_build/dumper

.PHONY: dump-all
dump-all: clean build
	./hack/dump-swagger-specs.sh

.PHONY: combine
combine: clean build
	_build/combiner

.PHONY: combine-db
combine-db: clean build
	_build/combiner > data/database.json

.PHONY: render
render: clean build
	_build/render

.PHONY: deploy
deploy:
	scp public/index.html kube-api.ninja:/srv/www/kube-api.ninja/www/
	scp public/about.html kube-api.ninja:/srv/www/kube-api.ninja/www/
	scp -r public/static/ kube-api.ninja:/srv/www/kube-api.ninja/www/

