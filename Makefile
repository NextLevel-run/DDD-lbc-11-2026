PNPM ?= pnpm
API := apps/api
WEB := apps/web

.DEFAULT_GOAL := help
.PHONY: help install dev dev-api dev-web test test-api test-web lint lint-api lint-web \
        build build-api build-web docs-budget check up down

help: ## List the targets
	@grep -E '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "} {printf "  %-12s %s\n", $$1, $$2}'

install: ## Download the API modules and the frontend dependencies
	cd $(API) && go mod download
	cd $(WEB) && $(PNPM) install --frozen-lockfile

dev: ## Run the API (:8080) and the frontend (:3000) together
	@$(MAKE) -j2 dev-api dev-web

dev-api: ## Run the API on :8080
	cd $(API) && go run ./cmd/api

dev-web: ## Run the frontend on :3000
	cd $(WEB) && $(PNPM) dev

test: test-api test-web ## Run the tests of both apps

test-api: ## Run the API tests
	cd $(API) && go test ./...

test-web: ## Run the frontend tests
	cd $(WEB) && $(PNPM) test

lint: lint-api lint-web ## Check format, lint and types of both apps

lint-api: ## gofmt + go vet
	@cd $(API) && unformatted="$$(gofmt -l .)"; \
	if [ -n "$$unformatted" ]; then echo "gofmt: run 'gofmt -w' on:"; echo "$$unformatted"; exit 1; fi
	cd $(API) && go vet ./...

lint-web: ## Prettier + ESLint + TypeScript
	cd $(WEB) && $(PNPM) check && $(PNPM) lint && $(PNPM) typecheck

build: build-api build-web ## Build both apps

build-api: ## Build the API binary to apps/api/bin/api
	cd $(API) && go build -o bin/api ./cmd/api

build-web: ## Build the frontend
	cd $(WEB) && $(PNPM) build

docs-budget: ## Check the size budget of the AGENTS.md files and the glossaries
	node tools/docs-budget/docs-budget.mjs
	node --test tools/docs-budget/docs-budget.test.mjs

check: docs-budget lint test build ## Everything CI runs

up: ## Run both apps in Docker Compose
	docker compose up --build

down: ## Stop Docker Compose
	docker compose down
