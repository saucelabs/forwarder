.PHONY: install-dependencies
install-dependencies:
	brew install hugo

.PHONY: gen
gen:
	@cd ../ && go run ./cmd/docsgen -docs-dir=$(PWD)

.PHONY: up
up:
	@hugo server --minify --disableFastRender

dist:
	@HUGO_ENVIRONMENT=production HUGO_ENV=production hugo --gc --minify
