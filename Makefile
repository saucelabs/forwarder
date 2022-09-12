# Copyright 2021 The forwarder Authors. All rights reserved.
# Use of this source code is governed by a MIT
# license that can be found in the LICENSE file.

all: dev

export GOBIN := $(PWD)/bin
export PATH  := $(GOBIN):$(PATH)

.PHONY: install-dependencies
install-dependencies:
	@rm -Rf bin && mkdir -p $(GOBIN)
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/godoc@latest

.PHONY: dev
dev:
	@air -c .air.toml

BUILD_BASE_PKG_NAME = github.com/saucelabs/forwarder/internal/
BUILD_GIT_COMMIT = $(shell git rev-list -1 HEAD)
BUILD_DATE = $(shell date)
BUILD_LDFLAGS = "-X '$(BUILD_BASE_PKG_NAME)version.buildCommit=$(BUILD_GIT_COMMIT)' -X '$(BUILD_BASE_PKG_NAME)version.buildVersion=$(BUILD_VERSION)' -X '$(BUILD_BASE_PKG_NAME)version.buildTime=$(BUILD_DATE)' -extldflags '-static'"

.PHONY: dev-build
dev-build:
	@go build -o ./forwarder.race -race -ldflags $(BUILD_LDFLAGS) ./.

.PHONY: lint
lint:
	@golangci-lint run -v -c .golangci.yml && echo "Lint OK"

.PHONY: test
test:
	@go test -timeout 120s -short -v -race -cover -coverprofile=coverage.out ./...

.PHONY: bench
bench:
	@go test -bench=. -run=XXX ./pkg/proxy # If you hit too many open files: ulimit -Sn 10000

.PHONY: integration-test
integration-test:
	@FORWARDER_TEST_MODE=integration go test -timeout 120s -v -race -cover -coverprofile=coverage.out ./...

.PHONY: coverage
coverage:
	@go tool cover -func=coverage.out

.PHONY: doc
doc:
	@echo "Open http://localhost:6060/pkg/github.com/saucelabs/forwarder/ in your browser\n"
	@godoc -http :6060
