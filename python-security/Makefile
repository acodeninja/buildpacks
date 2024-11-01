.DEFAULT_GOAL := help
SHELL := bash

PACKAGES := (python-security)

include ../scripts/build.mk
include ../scripts/requirements.mk
include ../scripts/package.mk

.PHONY: test
test: test-python-pip test-python-pipenv

.PHONY: test-python-pip
test-python-pip: package
	@cd fixtures/python-pip && pack build --clear-cache test-python-security-pip
	@docker run --rm --platform linux/amd64 --entrypoint test test-python-security-pip

.PHONY: test-python-pipenv
test-python-pipenv: package
	@cd fixtures/python-pipenv && pack build --clear-cache test-python-security-pipenv
	@docker run --rm --platform linux/amd64 --entrypoint test test-python-security-pipenv
