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
	@cd fixtures/python-pipenv && pack build test-app-pipenv
	@docker run --rm --platform linux/amd64 --entrypoint test test-app-pipenv

.PHONY: test-python-pip
test-python-pip: package
	@cd fixtures/python-pip && pack build test-app-pip
	@docker run --rm --platform linux/amd64 --entrypoint test test-app-pip

.PHONY: test-python-poetry
test-python-poetry: package
	@cd fixtures/python-poetry && pack build test-app-poetry
	@docker run --rm --platform linux/amd64 --entrypoint test test-app-poetry
