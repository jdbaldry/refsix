.ONESHELL:
.DELETE_ON_ERROR:
export SHELL     := bash
export SHELLOPTS := pipefail:errexit
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rule

# Adapted from https://www.thapaliya.com/en/writings/well-documented-makefiles/
.PHONY: help
help: ## Display this help.
help:
	@awk 'BEGIN {FS = ": ##"; printf "Usage:\n  make <target>\n\nTargets:\n"} /^[a-zA-Z0-9_\.\-\/%]+: ##/ { printf "  %-45s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

HTML_FILES = $(shell find -type f -name '*.html')

.PHONY: cleanup-html
cleanup-html: ## Clean up and format HTML files.
	./scripts/cleanup-html $(HTML_FILES)

stats.html: ## Generate the statistics page.
stats.html: stats.html.tmpl main.go events.go
	go run ./ > $@
