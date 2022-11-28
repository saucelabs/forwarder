# Ignore CPU pinning in CI.
ifdef CI
COMPOSE := docker compose -f docker-compose.yaml $(foreach f,$(CONF),-f $(f))
else
COMPOSE := docker compose -f docker-compose.yaml -f override/cpuset.yaml -f override/proxy-expose-ports.yaml $(foreach f,$(CONF),-f $(f))
endif

# Docker tag to be tested.
ifndef FORWARDER_VERSION
export FORWARDER_VERSION := devel
endif

.PHONY: up
up:
	@echo ">>> Using $(CONF)"
	@$(COMPOSE) up -d --wait --force-recreate --remove-orphans

.PHONY: down
down:
	@$(COMPOSE) down -v --remove-orphans

.PHONY: term
term:
	@docker kill -s TERM forwarder-e2e-proxy-1

.PHONY: dump-logs
dump-logs:
	@$(COMPOSE) logs $(SRV)

.PHONY: test
test: RUN ?= .
test: e2e.test
	@docker run --name "test-runner" --network "forwarder-e2e_default" --cpuset-cpus 0 \
	-v "$(PWD)/e2e.test:/usr/bin/e2e.test" -i --read-only --rm \
	$(shell docker inspect forwarder-e2e-proxy-1 --format='{{ range .Config.Env }}{{ if eq "FORWARDER" (slice . 0 9) }} -e {{.}}{{end}}{{end}}') \
	ubuntu e2e.test -test.run $(RUN) $(ARGS)

e2e.test: *.go
	@CGO_ENABLED=0 GOOS=linux go test -tags e2e -c -o e2e.test .
