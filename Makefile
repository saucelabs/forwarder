export GOBIN ?= $(CURDIR)/bin
export PATH  := $(GOBIN):$(PATH)

include .version

ifneq ($(shell expr $(MAKE_VERSION) \>= 4), 1)
$(error This Makefile requires GNU Make version 4 or higher, got $(MAKE_VERSION))
endif

ifneq ($(GO_VERSION),$(shell go version | grep -o -E '1\.[0-9\.]+'))
$(error Go version $(GO_VERSION) is required, got $(shell go version))
endif

.PHONY: install-dependencies
install-dependencies:
	@rm -Rf bin && mkdir -p $(GOBIN)
	go install golang.org/x/tools/cmd/godoc@$(X_TOOLS_VERSION)
	go install golang.org/x/tools/cmd/stringer@$(X_TOOLS_VERSION)
	go install golang.org/x/tools/cmd/stress@$(X_TOOLS_VERSION)

	go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	go install github.com/goreleaser/goreleaser/v2@$(GORELEASER_VERSION)
	go install github.com/google/go-licenses@$(GO_LICENSES_VERSION)

.PHONY: build
build:
	@rm -f forwarder
	@goreleaser build --clean --snapshot --single-target --output .

.PHONY: dist
dist: GORELEASER_CURRENT_TAG=1.0.0-rc
dist:
	@goreleaser --clean --snapshot --skip=docker,publish

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

export BUILDAH_FORMAT=docker

GOBUILD = go build -tags dnshack -ldflags "-checklinkname=0"

.PHONY: update-devel-image
update-devel-image: CONTAINER_RUNTIME ?= docker
update-devel-image: TAG=devel
update-devel-image: TMPDIR:=$(shell mktemp -d)
update-devel-image:
	@ln Containerfile LICENSE LICENSE.3RD_PARTY $(TMPDIR)
ifeq ($(shell uname),Linux)
	@CGO_ENABLED=1 GOOS=linux $(GOBUILD) -race -o $(TMPDIR)/forwarder ./cmd/forwarder
	@$(CONTAINER_RUNTIME) buildx build --network host -f Containerfile --build-arg BASE_IMAGE=ubuntu:latest -t saucelabs/forwarder:$(TAG) $(TMPDIR)
else
	@CGO_ENABLED=0 GOOS=linux $(GOBUILD) -o $(TMPDIR)/forwarder ./cmd/forwarder
	@$(CONTAINER_RUNTIME) buildx build --network host -f Containerfile -t saucelabs/forwarder:$(TAG) $(TMPDIR)
endif
	@rm -rf $(TMPDIR)

LICENSE.3RD_PARTY: LICENSE.3RD_PARTY.tpl go.mod go.sum
	@go-licenses report ./cmd/forwarder --template LICENSE.3RD_PARTY.tpl --ignore $(shell go list .) --ignore golang.org > LICENSE.3RD_PARTY

.forwarder.yaml:
	@cp .forwarder.yaml.example .forwarder.yaml

.PHONY: run
run: .forwarder.yaml
run:
	@GOMAXPROCS=1 go run ./cmd/forwarder run --config-file .forwarder.yaml

.PHONY: run-race
run-race: .forwarder.yaml
run-race:
	@go run --race ./cmd/forwarder run --config-file .forwarder.yaml

.PHONY: cpu.prof
cpu.prof:
	@curl -s http://localhost:10000/debug/pprof/profile?seconds=30 > cpu.prof

.PHONY: mem.prof
mem.prof:
	@curl -s http://localhost:10000/debug/pprof/heap > mem.prof

.PHONY: mutex.prof
mutex.prof:
	@curl -s http://localhost:10000/debug/pprof/mutex > mutex.prof
