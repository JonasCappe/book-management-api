.PHONY: help run test db-up db-down db-init db-logs db-config

DIRENV := direnv exec .
COMPOSE := $(DIRENV) docker compose

help:
	@echo "Available targets:"
	@echo "  make run        Run the API with the direnv environment"
	@echo "  make test       Run all Go tests with the direnv environment"
	@echo "  make db-up      Start PostgreSQL"
	@echo "  make db-down    Stop PostgreSQL"
	@echo "  make db-init    Apply the database schema to the running database"
	@echo "  make db-logs    Follow PostgreSQL logs"
	@echo "  make db-config  Render and validate the Compose configuration"

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
