export FORWARDER_VERSION ?= devel

CONTAINER_RUNTIME ?= docker
COMPOSE := $(CONTAINER_RUNTIME) compose -f compose.yaml

.PHONY: update-test-image
update-test-image:
	@CGO_ENABLED=0 GOOS=linux go test -tags e2e -c -o e2e.test ./tests
	@$(CONTAINER_RUNTIME) buildx build --network host -t forwarder-e2e -f e2e.Containerfile .

.PHONY: update-helper-images
update-helper-images:
	@$(MAKE) -C dns update-image
	@$(MAKE) -C sc-2450 update-image

.PHONY: run-e2e
run-e2e: SETUP ?= .
run-e2e: PARALLEL ?= 2
run-e2e:
	@go run . -setup "$(SETUP)" -run "$(RUN)" -parallel $(PARALLEL) $(ARGS)
