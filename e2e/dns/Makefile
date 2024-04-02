CONTAINER_RUNTIME ?= docker

export BUILDAH_FORMAT=docker

.PHONY: update-image
update-image:
	@$(CONTAINER_RUNTIME) buildx build --network host -t e2e-dns -f Containerfile .
