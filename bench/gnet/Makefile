.PHONY: run
run:
	@go run .

.PHONY: bench
bench:
	wrk http://localhost:8080 -d 10 -t 1 -c 100

GIT_ROOT := $(shell git rev-parse --show-toplevel)

.PHONY: run-forwarder
run-forwarder:
	@cd $(GIT_ROOT) && CGO_ENABLED=0 go run -trimpath -ldflags "-extldflags '-static'" ./cmd/forwarder run --config-file $(PWD)/forwarder.yaml

.PHONY: run-envoy
run-envoy:
	@envoy -c envoy.yaml

.PHONY: bench-proxy
bench-proxy:
	wrk http://localhost:3128 -d 10 -t 1 -c 100
