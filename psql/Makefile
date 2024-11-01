.DEFAULT_GOAL := help
SHELL := bash

PACKAGES := (psql)

include ../scripts/build.mk
include ../scripts/requirements.mk
include ../scripts/package.mk

.PHONY: test
test: test-bash

.PHONY: test-bash
test-bash: package
	@cd fixtures/bash && pack build --clear-cache test-psql-bash
	@docker run --rm --platform linux/amd64 --entrypoint test test-psql-bash
