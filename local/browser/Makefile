all: install test

venv:
	python -m venv venv

export PATH  := $(PWD)/venv/bin:$(PATH)

.PHONY: install
install: venv
	pip install -r requirements.txt
	playwright install firefox

.PHONY: test
test:
	@python ./topgun.py
