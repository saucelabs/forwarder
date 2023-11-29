.PHONY: install-dependencies
install-dependencies:
	brew install hugo

.PHONY: gen
gen:
	@cd ../ && go run ./cmd/docsgen -docs-dir=$(PWD)

.PHONY: up
up:
	@hugo server --minify --disableFastRender

.PHONY: dist
dist: BASE_URL="https://saucelabs.github.io/forwarder/"
dist:
	@HUGO_ENVIRONMENT=production HUGO_ENV=production hugo --gc --minify --baseURL "$(BASE_URL)"
