# Ignore CPU pinning in CI.
ifdef CI
CPUSET := ""
else
CPUSET := -f override/cpuset.yaml
endif

# Docker tag to be tested.
ifndef FORWARDER_VERSION
export FORWARDER_VERSION := devel
endif

.PHONY: up
up:
	@docker-compose -f docker-compose.yaml $(CPUSET) $(foreach f,$(CONF),-f $(f) ) \
	up -d --wait --force-recreate --remove-orphans

.PHONY: down
down:
	@docker-compose down -v --remove-orphans

.PHONY: dump-logs
dump-logs:
	@docker-compose logs $(SRV)

.PHONY: test
test: RUN=.
test: e2e.test
	@docker run --name "test-runner" --network "forwarder-e2e_default" --cpuset-cpus 0 \
	-v "$(PWD)/e2e.test:/usr/bin/e2e.test" -i --read-only --rm ubuntu \
	e2e.test -test.run $(RUN) $(ARGS)

e2e.test: *.go
	@CGO_ENABLED=0 GOOS=linux go test -tags e2e -c -o e2e.test .
