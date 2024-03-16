SHELL := bash

.PHONY: build
build: setup
	@cd app-$$PACKAGE_MANAGER && pack build test-app-$$PACKAGE_MANAGER

.PHONY: test
test: build
	@docker run --platform linux/amd64 --entrypoint test test-app-$$PACKAGE_MANAGER

.PHONY: setup
setup:
	@cp app/* app-pip/
	@cp app/* app-pipenv/
	@cp app/* app-poetry/
