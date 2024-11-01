.DEFAULT_GOAL := help
SHELL := bash

PACKAGES := (playwright)

include ../scripts/build.mk
include ../scripts/requirements.mk
include ../scripts/package.mk

.PHONY: test
test: test-python-pipenv test-python-pip test-python-poetry

.PHONY: test-python-pipenv
test-python-pipenv: package
	@cd fixtures/python-pipenv && pack build --clear-cache test-playwright-pipenv
	@docker run --rm --platform linux/amd64 --entrypoint test test-playwright-pipenv

.PHONY: test-python-pip
test-python-pip: package
	@cd fixtures/python-pip && pack build --clear-cache test-playwright-pip
	@docker run --rm --platform linux/amd64 --entrypoint test test-playwright-pip

.PHONY: test-python-poetry
test-python-poetry: package
	@cd fixtures/python-poetry && pack build --clear-cache test-playwright-poetry
	@docker run --rm --platform linux/amd64 --entrypoint test test-playwright-poetry
