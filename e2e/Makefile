COMPOSE := docker compose -f docker-compose.yaml

# Docker tag to be tested.
ifndef FORWARDER_VERSION
export FORWARDER_VERSION := devel
endif

E2E_TEST_IMAGE ?= gcr.io/distroless/static:nonroot

.PHONY: run-e2e
run-e2e: SETUP ?= .
run-e2e:
	@docker pull $(E2E_TEST_IMAGE) > /dev/null 2> /dev/null || true # Best effort to pre-pull image.
	@go run . -setup "$(SETUP)" $(SETUP_ARGS)

.PHONY: up
up:
	@$(COMPOSE) up -d --wait --force-recreate --remove-orphans

.PHONY: down
down:
	@$(COMPOSE) down -v --remove-orphans

.PHONY: term
term: SRV ?= forwarder-e2e-proxy-1
term:
	@docker kill -s TERM $(SRV)

.PHONY: dump-logs
dump-logs:
	@$(COMPOSE) logs $(SRV)

.PHONY: test
test: RUN ?= .
test: e2e.test
	@docker run --name "test-runner" --network "forwarder-e2e_default" --cpuset-cpus 0 \
	-v "$(CURDIR)/e2e.test:/usr/bin/e2e.test" \
	-v "$(CURDIR)/certs/ca.crt:/etc/forwarder/certs/ca.crt:ro" \
	-i --read-only --rm \
	$(shell docker inspect forwarder-e2e-proxy-1 --format='{{ range .Config.Env }}{{ if eq "FORWARDER" (slice . 0 9) }} -e "{{.}}"{{end}}{{end}}') \
	$(shell docker inspect forwarder-e2e-httpbin-1 --format='{{ range .Config.Env }}{{ if eq "FORWARDER" (slice . 0 9) }} -e "HTTPBIN{{slice . 9}}"{{end}}{{end}}' 2> /dev/null) \
	$(E2E_TEST_IMAGE) e2e.test -test.run "$(RUN)" -test.shuffle on -test.v $(ARGS)

.PHONY: e2e.test
e2e.test:
	@CGO_ENABLED=0 GOOS=linux go test -tags e2e -c -o e2e.test ./tests

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
