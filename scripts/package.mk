.PHONY: package
package: build ## Compile the buildpack.
	@echo "Packaging buildpacks... "
	@declare -a PARSED_PACKAGES=$(PACKAGES) ; \
	for PACKAGE in "$${PARSED_PACKAGES[@]}" ; do \
	  	CURRENT_DIR=$(PWD); \
	  	cd "$(GIT_ROOT)/$$PACKAGE"; \
  		echo -n "  $$PACKAGE "; \
		(pack buildpack package "ghcr.io/`cat $(GIT_ROOT)/$$PACKAGE/buildpack.toml | yj -t | jq -rc '.buildpack.id + ":" + .buildpack.version'`" > /dev/null && \
  		echo -e "\033[36m✔\033[0m") || \
  		echo -e "\033[31m✗\033[0m"; \
  		cd $(CURRENT_DIR); \
  	done
