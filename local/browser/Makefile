all: install test

.PHONY: clean
clean:
	rm -rf venv

venv:
	python -m venv venv

export PATH  := $(PWD)/venv/bin:$(PATH)

.PHONY: install
install: venv
	pip install -r requirements.txt
	playwright install chromium

.PHONY: test
test:
	@python ./topgun.py
