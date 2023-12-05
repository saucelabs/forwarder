.PHONY: install-dependencies
install-dependencies:
	brew install hugo

.PHONY: gen
gen:
	@cd ../ && go run ./cmd/docsgen -docs-dir=$(PWD)

.PHONY: favicon
favicon:
	@convert ./static/sl.svg -resize 32x32 ./static/icons/favicon-32x32.png
	@convert ./static/sl.svg -resize 128x128 ./static/icons/favicon-128x128.png
	@convert ./static/sl.svg -resize 180x180 ./static/icons/favicon-180x180.png
	@convert ./static/sl.svg -resize 192x192 ./static/icons/favicon-192x192.png
	@cp ./static/favicon-128x128.png ./static/favicon.png

.PHONY: up
up:
	@hugo server --minify --disableFastRender

.PHONY: dist
dist: BASE_URL="https://saucelabs.github.io/forwarder/"
dist:
	@HUGO_ENVIRONMENT=production HUGO_ENV=production hugo --gc --minify --baseURL "$(BASE_URL)"
