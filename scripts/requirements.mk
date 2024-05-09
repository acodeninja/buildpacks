.PHONY: has-requirements
has-requirements:
	@echo "Checking requirements... "
	@(docker ps > /dev/null && echo -e "  docker \033[36m✔\033[0m ") || (echo "Docker cli not installed or set up" && exit 1)
	@(command -v pack > /dev/null && echo -e "  pack \033[36m✔\033[0m ") || (echo "Pack cli not installed" && exit 1)
	@(command -v go > /dev/null && echo -e "  golang \033[36m✔\033[0m ") || (echo "Golang not installed" && exit 1)
	@(command -v yj > /dev/null && echo -e "  yj \033[36m✔\033[0m ") || (echo "YJ not installed" && exit 1)
	@(command -v jq > /dev/null && echo -e "  jq \033[36m✔\033[0m ") || (echo "JQ not installed" && exit 1)
	@(command -v git > /dev/null && echo -e "  git \033[36m✔\033[0m ") || (echo "git not installed" && exit 1)
	@(command -v sed > /dev/null && echo -e "  sed \033[36m✔\033[0m ") || (echo "sed not installed" && exit 1)
