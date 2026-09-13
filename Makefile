include .envrc
.PHONY: help run test db-up db-down db-init db-logs db-config migrate-create migrate-up migrate-down migrate-version migrate-force

DIRENV := direnv exec .
COMPOSE := $(DIRENV) docker compose
MIGRATIONS_PATH := ./cmd/migrate/migrations

help:
	@echo "Available targets:"
	@echo "  make run        Run the API with the direnv environment"
	@echo "  make test       Run all Go tests with the direnv environment"
	@echo "  make db-up      Start PostgreSQL"
	@echo "  make db-down    Stop PostgreSQL"
	@echo "  make db-init    Apply the database schema to the running database"
	@echo "  make db-logs    Follow PostgreSQL logs"
	@echo "  make db-config  Render and validate the Compose configuration"
	@echo "  make migrate-up Apply all pending migrations"
	@echo "  make migrate-down Roll back the latest migration"
	@echo "  make migrate-version Show the current migration version"
	@echo "  make migrate-create name=<name> Create a migration pair"
	@echo "  make migrate-force version=<n> Force a dirty database version"

run:
	$(DIRENV) go run ./cmd/api

test:
	$(DIRENV) go test ./...

db-up:
	$(COMPOSE) up -d postgres

db-down:
	$(COMPOSE) down

db-init:
	$(DIRENV) sh -c 'docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U "$$POSTGRES_USER" -d "$$POSTGRES_DB" < scripts/db_init.sql'

db-logs:
	$(COMPOSE) logs -f postgres

db-config:
	$(COMPOSE) config

migrate-create:
	@test -n "$(name)" || (echo "usage: make migrate-create name=<name>" && exit 1)
	$(DIRENV) migrate create -seq -ext .sql -dir $(MIGRATIONS_PATH) "$(name)"

migrate-up:
	$(DIRENV) sh -c 'migrate -path "$(MIGRATIONS_PATH)" -database "$$DB_DSN" up'

migrate-down:
	$(DIRENV) sh -c 'migrate -path "$(MIGRATIONS_PATH)" -database "$$DB_DSN" down 1'

migrate-version:
	$(DIRENV) sh -c 'migrate -path "$(MIGRATIONS_PATH)" -database "$$DB_DSN" version'

migrate-force:
	@test -n "$(version)" || (echo "usage: make migrate-force version=<n>" && exit 1)
	$(DIRENV) sh -c 'migrate -path "$(MIGRATIONS_PATH)" -database "$$DB_DSN" force "$(version)"'

.PHONY: gen-docs
gen-docs:
	swag fmt -d cmd/api
	swag init -g main.go -d cmd/api,internal/data -o docs
