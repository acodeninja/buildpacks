.DEFAULT_GOAL := help
SHELL := bash

PACKAGES := (graphviz)

include ../scripts/build.mk
include ../scripts/requirements.mk
include ../scripts/package.mk

.PHONY: test
test: test-basic

.PHONY: test-basic
test-basic: package
	@cd fixtures/basic && pack build --clear-cache test-basic
	@docker run --rm --platform linux/amd64 --entrypoint test test-basic
