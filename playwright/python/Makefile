SHELL := bash

.PHONY: build
build:
	@cd app-$$PACKAGE_MANAGER && pack build test-app-$$PACKAGE_MANAGER

.PHONY: test
test: build
	@docker run --platform linux/amd64 test-app-$$PACKAGE_MANAGER
