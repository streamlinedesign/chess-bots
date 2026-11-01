SHELL := bash
.DEFAULT_GOAL := help

# ---- Project ----------------------------------------------------------------
# chess-bots: Go command line (bots and a simulator for Eat the King). No
# server, so no port (PORTS.md: "command line, no port"). .env overrides.
NAME     ?= chess-bots
KIND     ?= cli
DROPLET  ?=
# -----------------------------------------------------------------------------

-include .env
export

VERSION := $(shell cat VERSION 2>/dev/null || echo 0.0.0)
NOT_USED = @echo "make $@: not used in this project"

help: ## List targets
	@grep -E '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk -F ':.*## ' '{printf "  %-14s %s\n", $$1, $$2}'

dev: ## Not used; run make simulate
	@echo "make dev: not used in this project; run make simulate"

# Plays bot games until Ctrl-C, adding each to games.jsonl, then prints a
# summary. Ctrl-C finishes the games in hand first; run it again to add more.
# Options go in ARGS, e.g. make simulate ARGS="-games 10000 -fog on".
# Every option: go run ./cmd/simulate -h
simulate: ## Play bot games into games.jsonl until Ctrl-C (options in ARGS)
	go run ./cmd/simulate -games 0 $(ARGS)

# The summary of everything in games.jsonl so far.
summary: ## Summarise everything in games.jsonl so far
	go run ./cmd/simulate -summary games.jsonl

test: ## All tests
	go test ./...

build: ## Build ./cmd/simulate to ./chess-bots
	go build -o $(NAME) ./cmd/simulate

lint: ## gofmt and go vet
	gofmt -l -w $$(git ls-files '*.go')
	go vet ./...

db-reset: ## Drop and recreate the local database, run migrations
	$(NOT_USED)

db-seed: ## Load demo data from seed/
	$(NOT_USED)

psql: ## Shell on the local database
	$(NOT_USED)

deploy: ## Build for linux/amd64 and ship to DROPLET
	$(NOT_USED)

release: ## Tag VERSION, move Unreleased in CHANGELOG.md, push the tag
	@test "$$(git branch --show-current)" = main || { echo "release from main only"; exit 1; }
	@git diff --quiet || { echo "commit first"; exit 1; }
	@grep -q "## \[Unreleased\]" CHANGELOG.md || { echo "CHANGELOG.md needs an Unreleased section"; exit 1; }
	sed -i.bak "s/## \[Unreleased\]/## [Unreleased]\n\n## [$(VERSION)] - $$(date +%Y-%m-%d)/" CHANGELOG.md && rm CHANGELOG.md.bak
	git commit -am "Release $(VERSION)"
	git tag -a v$(VERSION) -m "v$(VERSION)"
	git push && git push --tags

configure-repo: ## One-time GitHub settings: auto-merge, Actions may open PRs, main protected, hold label
	@repo=$$(gh repo view --json nameWithOwner --jq .nameWithOwner); \
	gh api -X PATCH repos/$$repo -F allow_auto_merge=true -F delete_branch_on_merge=true -F allow_squash_merge=true >/dev/null; \
	gh api -X PUT repos/$$repo/actions/permissions/workflow -F default_workflow_permissions=write -F can_approve_pull_request_reviews=true >/dev/null; \
	gh api -X PUT repos/$$repo/branches/main/protection --input - >/dev/null <<< '{"required_status_checks":{"strict":false,"contexts":["ci"]},"enforce_admins":false,"required_pull_request_reviews":null,"restrictions":null,"allow_force_pushes":false,"allow_deletions":false}'; \
	gh label create hold --color B60205 --description "Do not auto-merge" --force >/dev/null; \
	echo "$$repo configured"

clean: ## Remove the built binary
	rm -f $(NAME) simulate

.PHONY: help dev simulate summary test build lint db-reset db-seed psql deploy release configure-repo clean
