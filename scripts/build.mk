GIT_ROOT := $(shell git rev-parse --show-toplevel)

.PHONY: build
build: has-requirements ## Compile the buildpack.
	@echo "Compiling binaries... "
	@declare -a PARSED_PACKAGES=$(PACKAGES) ; \
	for PACKAGE in "$${PARSED_PACKAGES[@]}" ; do \
  		echo -n "  $$PACKAGE "; \
	  	CURRENT_DIR=$(PWD); \
	  	cd "$(GIT_ROOT)/$$PACKAGE"; \
	  	rm -f ./bin/*; \
		(GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/main "github.com/acodeninja/buildpacks/$$PACKAGE" && \
  		echo -e "\033[36m✔\033[0m") || \
  		echo -e "\033[31m✗\033[0m"; \
  		cp bin/main bin/detect; \
  		cp bin/main bin/build; \
  		cd $(CURRENT_DIR); \
  	done
