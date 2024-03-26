CONTAINER_RUNTIME ?= docker
COMPOSE := $(CONTAINER_RUNTIME) compose -f compose.yaml

.PHONY: up
up:
	@$(COMPOSE) up -d --wait --force-recreate --remove-orphans
	@echo "Dashboard: http://localhost:3000/d/daf42347-6322-401c-af67-e30c59d21d8d/forwarder"

.PHONY: down
down:
	@$(COMPOSE) down -v --remove-orphans
