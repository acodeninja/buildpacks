.DEFAULT_GOAL := help
SHELL := bash

PACKAGES := (install)

include ../scripts/build.mk
include ../scripts/requirements.mk
include ../scripts/package.mk

.PHONY: test
test: test-apt

.PHONY: test-apt
test-apt: package
	@cd fixtures/apt && pack build test-install
	@docker run --rm --platform linux/amd64 --entrypoint test test-install
