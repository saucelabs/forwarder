export COMPOSE_PROJECT_NAME = forwarder-e2e

CONTAINER_RUNTIME ?= docker
COMPOSE := $(CONTAINER_RUNTIME) compose -f compose.yaml

# Docker tag to be tested.
ifndef FORWARDER_VERSION
export FORWARDER_VERSION := devel
endif

E2E_TEST_IMAGE ?= gcr.io/distroless/static:nonroot

.PHONY: update-helper-images
update-helper-images:
	@$(MAKE) -C dns update-image
	@$(MAKE) -C sc-2450 update-image

.PHONY: update-test-image
update-test-image:
	@CGO_ENABLED=0 GOOS=linux go test -tags e2e -c -o e2e.test ./tests
	@$(CONTAINER_RUNTIME) buildx build --network host -t forwarder-e2e -f e2e.Containerfile .

.PHONY: run-e2e
run-e2e: SETUP ?= .
run-e2e:
	@go run . -setup "$(SETUP)" $(SETUP_ARGS)

.PHONY: up
up:
	@$(COMPOSE) up -d --wait --force-recreate --remove-orphans

.PHONY: down
down:
	@$(COMPOSE) down -v --remove-orphans

define container_name
$(COMPOSE_PROJECT_NAME)-$(1)-1
endef

.PHONY: term
term: SRV ?= proxy
term:
	@$(CONTAINER_RUNTIME) kill -s TERM $(call container_name,$(SRV))

.PHONY: dump-logs
dump-logs:
	@$(COMPOSE) logs

.PHONY: test
test: RUN ?= .
test: e2e.test
	@$(CONTAINER_RUNTIME) run --name "test-runner" --network "$(COMPOSE_PROJECT_NAME)_default" --cpuset-cpus 0 \
	-v "$(CURDIR)/e2e.test:/usr/bin/e2e.test" \
	-v "$(CURDIR)/certs/ca.crt:/etc/forwarder/certs/ca.crt:ro" \
	-i --read-only --rm \
	$(shell $(CONTAINER_RUNTIME) inspect $(call container_name,proxy)   --format='{{ range .Config.Env }}{{ if eq "FORWARDER" (slice . 0 9) }} -e "{{.}}"{{end}}{{end}}') \
	$(shell $(CONTAINER_RUNTIME) inspect $(call container_name,httpbin) --format='{{ range .Config.Env }}{{ if eq "FORWARDER" (slice . 0 9) }} -e "HTTPBIN{{slice . 9}}"{{end}}{{end}}' 2> /dev/null) \
	$(E2E_TEST_IMAGE) e2e.test -test.run "$(RUN)" -test.shuffle on -test.v $(ARGS)

.PHONY: bench
bench: PROFILE ?= 10
bench: RUN ?= .
bench: TMPDIR := $(shell mktemp -d -t "com.saucelabs.Forwarder.XXXXXX")
bench:
	@echo ">>> Profiling enabled output in ${TMPDIR}"
	@curl -sS "http://localhost:10000/debug/pprof/profile?seconds=${PROFILE}" -o "${TMPDIR}/cpu" &
	@ARGS="-test.bench ${RUN} -test.benchtime ${PROFILE}s" RUN=^XXX make test
	@curl -sS "http://localhost:10000/debug/pprof/allocs" -o "${TMPDIR}/allocs"
	@curl -sS "http://localhost:10000/debug/pprof/heap" -o "${TMPDIR}/heap"
	@curl -sS "http://localhost:10000/debug/pprof/mutex" -o "${TMPDIR}/mutex"
