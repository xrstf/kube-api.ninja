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
	go build $(GO_BUILD_FLAGS) -o $(OUTPUT_DIR)/ ./cmd/clusterdumper
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
full-rebuild: make clean build dump-swagger render

.PHONY: dump-swagger
dump-swagger:
	./hack/dump-swagger-specs.sh

.PHONY: render
render:
	ASSET_STAMP=$(GIT_HEAD) _build/render

.PHONY: deploy
deploy:
	rsync --delete --recursive public/ xrstf@kube-api.ninja:/srv/www/kube-api.ninja/www

.PHONY: build-refdocs-image
build-refdocs-image:
	docker build --no-cache -t kubernetes-apidocs:latest hack/containers/kubernetes-reference-docs/
