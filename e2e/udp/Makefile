CONTAINER_RUNTIME ?= docker

.PHONY: server
server:
	@CGO_ENABLED=0 GOOS=linux go build ./cmd/server

export BUILDAH_FORMAT=docker

.PHONY: update-image
update-image: server
	@$(CONTAINER_RUNTIME) buildx build --network host -t e2e-udp -f Containerfile .
