# SPDX-FileCopyrightText: 2023 Dyne.org foundation
#
# SPDX-License-Identifier: AGPL-3.0-or-later

ANSIPLAY = ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook --inventory hosts.yaml --ssh-common-args '-o StrictHostKeyChecking=accept-new -o IdentitiesOnly=yes' --private-key ./sshkey $(1)

##@ General
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' Makefile

##@ Server management
install: ## 🔨 Install controller
	$(call ANSIPLAY, install-proxy.yml)

tls: ## 🔐 Challenge for SSL/TLS
	$(call ANSIPLAY, acme-certificate.yml)
