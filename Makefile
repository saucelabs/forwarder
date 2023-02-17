# Copyright 2021 The forwarder Authors. All rights reserved.
# Use of this source code is governed by a MIT
# license that can be found in the LICENSE file.

export GOBIN ?= $(PWD)/bin
export PATH  := $(GOBIN):$(PATH)

include .version

.PHONY: install-dependencies
install-dependencies:
	@rm -Rf bin && mkdir -p $(GOBIN)
	go install golang.org/x/tools/cmd/godoc@$(GODOC_VERSION)
	go install golang.org/x/tools/cmd/stringer@$(STRINGER_VERSION)

	go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	go install github.com/goreleaser/goreleaser@$(GORELEASER_VERSION)
	go install github.com/google/go-licenses@$(GO_LICENSES_VERSION)

.PHONY: build
build:
	@goreleaser build --rm-dist --snapshot --single-target --output .

.PHONY: dist
dist: GORELEASER_CURRENT_TAG=1.0.0-rc
dist:
	@goreleaser --clean --snapshot

.PHONY: docs
docs:
	@echo "Open http://localhost:6060/pkg/github.com/saucelabs/forwarder/ in your browser\n"
	@godoc -http :6060

.PHONY: clean
clean:
	@rm -Rf bin dist *.coverprofile *.dev *.race *.test *.log
	@go clean -cache -modcache -testcache ./... ||:

### Development and testing

.PHONY: .check-go-version
.check-go-version:
	@[[ "`go version`" =~ $(GO_VERSION) ]] || echo "[WARNING] Required Go version $(GO_VERSION) found `go version | grep -o -E '1\.[0-9\.]+'`"

.PHONY: fmt
fmt:
	@golangci-lint run -c .golangci-fmt.yml --fix ./...

.PHONY: lint
lint:
	@LOG_LEVEL=error golangci-lint run

.PHONY: test
test:
	@go test -timeout 120s -short -race -cover -coverprofile=coverage.out ./...

.PHONY: coverage
coverage:
	@go tool cover -func=coverage.out

.PHONY: update-devel-image
update-devel-image: TAG=devel
update-devel-image: TMPDIR:=$(shell mktemp -d)
update-devel-image:
	@CGO_ENABLED=0 GOOS=linux go build -tags e2e -o $(TMPDIR)/forwarder ./cmd/forwarder
	@ln Dockerfile LICENSE LICENSE.3RD_PARTY $(TMPDIR)
	@docker buildx build -t saucelabs/forwarder:$(TAG) $(TMPDIR)
	@rm -rf $(TMPDIR)

LICENSE.3RD_PARTY: LICENSE.3RD_PARTY.tpl go.mod go.sum
	@go-licenses report ./cmd/forwarder --template LICENSE.3RD_PARTY.tpl --ignore $(shell go list .) --ignore golang.org > LICENSE.3RD_PARTY
