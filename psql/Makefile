.DEFAULT_GOAL := help
SHELL := bash

PACKAGES := (psql)

include ../scripts/build.mk
include ../scripts/requirements.mk
include ../scripts/package.mk

.PHONY: test
test: test-bash-default test-bash-file

.PHONY: test-bash-default
test-bash-default: package
	@cd fixtures/bash-default && pack build --clear-cache test-psql-bash
	@cd fixtures/bash-default && docker compose up --abort-on-container-exit

.PHONY: test-bash-file
test-bash-file: package
	@cd fixtures/bash-file && pack build --clear-cache test-psql-bash
	@cd fixtures/bash-file && docker compose up --abort-on-container-exit
