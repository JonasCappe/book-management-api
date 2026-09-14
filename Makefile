.PHONY: \
	help run test smoke seed \
	db-up db-down db-init db-logs db-config \
	migrate-create migrate-up migrate-down migrate-version migrate-force \
	docker-build docker-up docker-down docker-logs \
	gen-docs

DIRENV := direnv exec .
COMPOSE := $(DIRENV) docker compose
MIGRATIONS_PATH := ./cmd/migrate/migrations

help:
	@echo "Application:"
	@echo "  make run                         Run the API locally"
	@echo "  make test                        Run all Go tests"
	@echo ""
	@echo "Database:"
	@echo "  make db-up                       Start PostgreSQL and wait until healthy"
	@echo "  make db-down                     Stop PostgreSQL"
	@echo "  make db-init                     Apply database bootstrap SQL manually"
	@echo "  make db-logs                     Follow PostgreSQL logs"
	@echo "  make db-config                   Validate Docker Compose configuration"
	@echo "  make seed                        Insert development seed data"
	@echo ""
	@echo "Migrations:"
	@echo "  make migrate-up                  Apply pending migrations"
	@echo "  make migrate-down                Roll back the latest migration"
	@echo "  make migrate-version             Show the current migration version"
	@echo "  make migrate-create name=<name>  Create a migration pair"
	@echo "  make migrate-force version=<n>   Force a migration version"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build                Build the API image"
	@echo "  make docker-up                   Start PostgreSQL, migrate, and start the API"
	@echo "  make docker-down                 Stop the Docker environment"
	@echo "  make docker-logs                 Follow API logs"
	@echo ""
	@echo "Documentation:"
	@echo "  make gen-docs                    Regenerate Swagger documentation"
	@echo ""
	@echo "Testing:"
	@echo "  make test                        Run all Go tests"
	@echo "  make smoke                       Run the API end-to-end smoke test"

run:
	$(DIRENV) go run ./cmd/api

test:
	$(DIRENV) go test ./...

seed:
	$(DIRENV) sh -c 'docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U "$$POSTGRES_USER" -d "$$POSTGRES_DB" < scripts/seed.sql'

smoke:
	$(DIRENV) ./scripts/smoke.sh

db-up:
	$(COMPOSE) up -d --wait postgres

db-down:
	$(COMPOSE) stop postgres

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

docker-build:
	$(COMPOSE) build api

docker-up: db-up migrate-up
	$(COMPOSE) up -d --build api

docker-down:
	$(COMPOSE) down

docker-logs:
	$(COMPOSE) logs -f api

gen-docs:
	swag fmt -d cmd/api
	swag init -g main.go -d cmd/api,internal/data -o docs