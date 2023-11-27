.PHONY: install-dependencies
install-dependencies:
	brew install hugo

.PHONY: up
up:
	@hugo server --minify --disableFastRender

dist:
	@HUGO_ENVIRONMENT=production HUGO_ENV=production hugo --gc --minify
