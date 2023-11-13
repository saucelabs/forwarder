VERSION := 2.48.0-rc.2

.PHONY: install
install:
	curl -sL https://github.com/prometheus/prometheus/releases/download/v$(VERSION)/prometheus-$(VERSION).darwin-arm64.tar.gz | tar xz -C ./bin --strip-components=1

.PHONY: run
run:
	@./bin/prometheus --config.file=./prometheus.yml --web.listen-address=:9090
