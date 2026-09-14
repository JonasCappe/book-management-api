# Book Management API

A REST API written in Go for managing books and maintaining an audit history of changes made to them.

The service supports creating, retrieving, updating, listing, and deleting books, while recording a human-readable and auditable history for each mutation.

## Features

* Book CRUD operations
* Multiple authors per book
* Publication date support
* Soft deletion
* Change history for create, update, and delete operations
* Human-readable change descriptions
* Structured JSON snapshots of changes
* Pagination and filtering
* Deterministic result ordering
* Request validation
* Consistent API error responses
* Strict JSON request decoding
* PostgreSQL persistence
* Transactional book and history updates
* Database migrations
* OpenAPI / Swagger documentation
* Health endpoint
* HTTP request timeouts
* Graceful HTTP server shutdown
* Structured JSON logging
* Request ID correlation
* Multi-stage Docker image
* Non-root container runtime
* Development seed data
* Automated tests
* End-to-end API smoke test

## Tech Stack

* Go
* chi
* PostgreSQL 17
* `database/sql`
* `lib/pq`
* `go-playground/validator`
* `log/slog`
* Swagger / swaggo
* golang-migrate
* Docker / Docker Compose
* direnv
* Make
* Bash / curl for smoke testing

## Architecture

The application is intentionally kept relatively small and explicit.

```text
HTTP Request
     │
     ▼
Handlers / Transport
     │
     ├── JSON decoding
     ├── request validation
     ├── query/path parsing
     └── HTTP error mapping
     │
     ▼
Storage interfaces
     │
     ▼
PostgreSQL
```

Cross-cutting concerns such as JSON handling, validation, error responses, pagination, configuration, logging, and database access are separated into dedicated components.

The project avoids introducing abstractions that do not provide meaningful value for the current scope of this exercise.

## Project Structure

```text
.
├── cmd/
│   ├── api/                    # HTTP API entry point and handlers
│   └── migrate/
│       └── migrations/         # Database migrations
├── docs/                       # Generated Swagger documentation
├── internal/
│   ├── data/                   # Domain/data structures
│   ├── db/                     # Database connection setup
│   ├── env/                    # Environment configuration helpers
│   └── store/                  # PostgreSQL persistence
├── scripts/
│   ├── db_init.sql             # PostgreSQL bootstrap SQL
│   ├── seed.sql                # Development seed data
│   └── smoke.sh                # End-to-end API smoke test
├── Dockerfile                  # Multi-stage API image
├── docker-compose.yml          # Local PostgreSQL and API environment
├── .dockerignore
├── .env.example
├── .envrc.example
├── Makefile
└── README.md
```

## Data Model

A book contains:

* ID
* title
* description
* publication date
* one or more authors
* creation timestamp
* update timestamp

Books and authors use a many-to-many relationship.

Books are soft-deleted so their history can remain available without breaking referential integrity.

## Change History

Every book mutation creates a history entry.

A history entry contains:

```json
{
  "id": 42,
  "book_id": 12,
  "changed_at": "2026-09-14T09:30:00Z",
  "change_type": "updated",
  "description": "Title changed from \"The Hobbitt\" to \"The Hobbit\"",
  "changes": {}
}
```

Supported change types are:

```text
created
updated
deleted
```

### Human-readable history

Updates are compared against the previous persisted state.

For example:

```text
Title changed from "The Hobbitt" to "The Hobbit"
```

Multiple changes can be combined:

```text
Title changed from "Old title" to "New title"; Author "J.R.R. Tolkien" was added
```

Author changes are compared by author ID, so changes in ordering do not create false history events.

### Machine-readable history

The API also stores structured JSON snapshots.

Creation records the resulting state:

```json
{
  "after": {
    "...": "..."
  }
}
```

Updates record both states:

```json
{
  "before": {
    "...": "..."
  },
  "after": {
    "...": "..."
  }
}
```

Deletion records the final state before deletion:

```json
{
  "before": {
    "...": "..."
  }
}
```

## Transactional Consistency

Book mutations and their history entries are written in the same PostgreSQL transaction.

An update follows approximately this flow:

```text
BEGIN

SELECT book FOR UPDATE

validate referenced authors

apply requested changes

UPDATE book

replace author relationships

load resulting state

INSERT history entry

COMMIT
```

If any step fails, the transaction is rolled back.

This means an application-level book mutation cannot successfully commit without its associated history entry also being persisted.

Updates use `SELECT ... FOR UPDATE` to lock the affected book while its previous state is being read and the corresponding history entry is constructed.

Partial updates are applied to the locked database state inside the transaction. This prevents omitted PATCH fields from being overwritten by stale values from an earlier read.

## Requirements

For local development:

* Go
* Docker
* Docker Compose
* direnv
* golang-migrate
* Make
* curl
* Bash
* swag CLI, when regenerating API documentation

## Configuration

Create the local environment files:

```bash
cp .env.example .env
cp .envrc.example .envrc
direnv allow
```

The example configuration is intended for local development only.

The application supports configuration such as:

```env
ADDR=":8080"
EXTERNALURL="localhost:8080"
ENVIRONMENT="development"

DB_DSN="postgres://postgres:postgres@localhost:5432/book-management?sslmode=disable"

DB_MAX_OPEN_CONNS=50
DB_MAX_IDLE_CONNS=50
DB_MAX_IDLE_TIME="15m"

POSTGRES_DB="book-management"
POSTGRES_USER="postgres"
POSTGRES_PASSWORD="postgres"
```

The API listens on port `8080` by default, and PostgreSQL is exposed on `localhost:5432`.

Do not use the example database credentials in a production environment.

## Running Locally

### 1. Start PostgreSQL

```bash
make db-up
```

PostgreSQL 17 runs through Docker Compose.

The command waits until the database reports healthy before returning.

### 2. Apply migrations

```bash
make migrate-up
```

Check the current migration version with:

```bash
make migrate-version
```

### 3. Optionally seed development data

Books require one or more existing authors.

Development-only author data can be inserted with:

```bash
make seed
```

The seed operation is safe to run repeatedly and is kept separate from schema migrations.

### 4. Run the API

```bash
make run
```

The API is available at:

```text
http://localhost:8080
```

Swagger UI is available at:

```text
http://localhost:8080/swagger/index.html
```

## Running with Docker

The API and PostgreSQL can also be run together through Docker Compose.

```bash
make docker-up
```

This command:

1. starts PostgreSQL and waits until it is healthy;
2. applies pending database migrations;
3. builds the API image;
4. starts the API container.

The API is available at:

```text
http://localhost:8080
```

Swagger UI is available at:

```text
http://localhost:8080/swagger/index.html
```

To add optional development authors:

```bash
make seed
```

Follow the API logs with:

```bash
make docker-logs
```

Build the API image without starting it:

```bash
make docker-build
```

Stop the complete Docker environment with:

```bash
make docker-down
```

The runtime image uses a non-root user and contains only the compiled application and the minimal runtime environment required to execute it.

## Development Seed Data

Development seed data is stored in:

```text
scripts/seed.sql
```

Load it with:

```bash
make seed
```

The seed contains authors that can be referenced when creating books through the API.

Seed data is intentionally kept separate from migrations because it is development and testing data rather than part of the database schema.

The seed script is designed to be repeatable and does not create duplicate authors when run multiple times.

## Available Make Targets

Display all available development commands with:

```bash
make help
```

The Makefile provides commands for:

* running the API;
* running automated tests;
* managing PostgreSQL;
* creating and applying database migrations;
* inserting development seed data;
* building and managing the Docker environment;
* running the end-to-end smoke test;
* regenerating Swagger documentation.

Common commands include:

```bash
make run
make test
make smoke

make db-up
make db-down
make db-logs
make db-config
make seed

make migrate-up
make migrate-down
make migrate-version
make migrate-create name=add_example_column

make docker-build
make docker-up
make docker-down
make docker-logs

make gen-docs
```

## API

All book endpoints are versioned under:

```text
/v1
```

### Health

```http
GET /health
```

### Books

```http
POST   /v1/books
GET    /v1/books
GET    /v1/books/{bookID}
PATCH  /v1/books/{bookID}
DELETE /v1/books/{bookID}
```

### Book History

```http
GET /v1/books/{bookID}/history
```

## Creating a Book

Example:

```json
{
  "title": "The Hobbit",
  "description": "A fantasy novel.",
  "publication_date": "1937-09-21",
  "author_ids": [1]
}
```

The API requires:

* a non-blank title;
* a publication date;
* at least one existing author.

Duplicate author IDs are rejected by request validation.

A duplicate book title results in a `409 Conflict` response.

If using the development environment, run:

```bash
make seed
```

before creating books manually so that existing author IDs are available.

## Partial Updates

`PATCH` uses optional fields so omitted properties remain unchanged.

Example:

```json
{
  "title": "The Hobbit",
  "author_ids": [1, 2]
}
```

A request containing no update fields is rejected.

Partial changes are applied inside the same transaction that locks and updates the book. This prevents concurrent PATCH requests from unintentionally restoring stale values for fields they did not modify.

## Listing Books

The book collection supports pagination, filtering, and ordering.

Example:

```http
GET /v1/books?page=1&page_size=20&title=hobbit&order_by=title&order=asc
```

Supported query parameters:

| Parameter | Description |
| --- | --- |
| `page` | Page number, starting at 1 |
| `page_size` | Number of records per page, maximum 100 |
| `title` | Partial, case-insensitive title filter |
| `author_id` | Filter by author |
| `published_from` | Minimum publication date |
| `published_to` | Maximum publication date |
| `order_by` | Sort field |
| `order` | `asc` or `desc` |

Supported `order_by` values:

```text
id
title
publication_date
created_at
updated_at
```

Dynamic sort fields are mapped through an allowlist before being included in SQL.

Result ordering uses the book ID as a secondary sort key where necessary to keep pagination deterministic.

## History Queries

Book history supports pagination, filtering by change type, and ordering.

History entries remain available after a book has been soft-deleted.

The history endpoint supports:

* pagination;
* filtering by change type;
* ascending or descending ordering.

Each entry includes both a human-readable description and structured JSON describing the associated state change.

## JSON Request Handling

JSON request bodies are deliberately strict.

The API:

* limits request bodies to 1 MiB;
* rejects malformed JSON;
* rejects unknown fields;
* rejects incorrect value types;
* rejects empty request bodies;
* rejects requests containing multiple JSON objects.

This helps detect client mistakes rather than silently ignoring invalid input.

## Validation

Request validation uses `go-playground/validator`.

Semantically invalid request fields return `422 Unprocessable Entity` with field-specific errors.

For example:

```json
{
  "error": {
    "code": "validation_failed",
    "message": "request validation failed",
    "fields": {
      "title": "must not be blank"
    }
  }
}
```

Malformed JSON and invalid request syntax return `400 Bad Request`.

## Errors

Errors use a consistent JSON representation:

```json
{
  "error": {
    "code": "validation_failed",
    "message": "request validation failed",
    "fields": {
      "title": "must not be blank"
    }
  }
}
```

Error categories include:

```text
invalid_request
validation_failed
not_found
conflict
internal_error
```

Typical status mappings are:

| Status | Meaning |
| --- | --- |
| `400 Bad Request` | Malformed request syntax or invalid path/query input |
| `404 Not Found` | Requested resource does not exist |
| `409 Conflict` | Request conflicts with existing state, such as a duplicate title |
| `422 Unprocessable Entity` | Request is structurally valid but fails field or domain validation |
| `500 Internal Server Error` | Unexpected server-side failure |

Internal database or implementation errors are not exposed directly to API consumers.

## Logging

The service uses structured JSON logging through Go's `log/slog` package.

HTTP request logs include:

* request ID;
* method;
* path;
* response status;
* response size;
* request duration.

Application errors include the request ID where available so request and error events can be correlated.

Example:

```json
{
  "level": "INFO",
  "msg": "request completed",
  "request_id": "example-request-id",
  "method": "GET",
  "path": "/v1/books",
  "status": 200
}
```

## Graceful Shutdown

The HTTP server handles termination signals and performs graceful shutdown.

When receiving `SIGINT` or `SIGTERM`, the service:

1. stops accepting new connections;
2. allows active requests a bounded amount of time to complete;
3. shuts down the HTTP server;
4. allows deferred resources such as the database connection pool to close normally.

This behavior also applies when the Docker container is stopped.

## API Documentation

Interactive Swagger documentation is exposed at:

```text
/swagger/index.html
```

Generated OpenAPI files are stored in:

```text
docs/
```

Regenerate them with:

```bash
make gen-docs
```

The generated specification reflects the behavior implemented by the service and does not advertise authentication that the API does not currently implement.

## Database Migrations

Migrations are stored under:

```text
cmd/migrate/migrations/
```

Schema changes are applied incrementally rather than modifying an existing migration after it has been introduced.

This repository currently includes migrations covering:

* authors;
* books;
* history entries;
* soft deletion;
* publication date;
* many-to-many book authors.

Apply all pending migrations with:

```bash
make migrate-up
```

Roll back the latest migration with:

```bash
make migrate-down
```

Create a new migration pair with:

```bash
make migrate-create name=add_example_column
```

Development seed data is deliberately not included in migrations.

## Testing

### Unit and integration tests

Run the Go test suite with:

```bash
make test
```

Or directly:

```bash
go test ./...
```

Tests cover API helpers, request and query validation, history behavior, and persistence-related functionality.

### End-to-end smoke test

The repository also contains a lightweight end-to-end smoke test:

```text
scripts/smoke.sh
```

The smoke test exercises the running application through its HTTP API rather than calling handlers or storage methods directly.

It verifies the primary workflow:

```text
health check
    ↓
create book
    ↓
retrieve book
    ↓
update book
    ↓
retrieve history
    ↓
list/filter books
    ↓
delete book
    ↓
verify book returns 404
    ↓
verify history remains available
```

Start the Docker environment and load the development authors:

```bash
make docker-up
make seed
```

Then run:

```bash
make smoke
```

A successful run ends with:

```text
Smoke test passed.
```

The smoke test is intentionally focused on verifying that the fully assembled application can execute its critical workflow.

Detailed validation and edge-case behavior remain covered by the Go test suite rather than duplicating those assertions in the shell smoke test.

A complete local verification can therefore be run with:

```bash
make test
make docker-up
make seed
make smoke
make docker-down
```

## Design Decisions

### History in the application layer

History generation is implemented in Go rather than through PostgreSQL triggers.

The API owns the book mutation path, so keeping this behavior in the application makes it explicit, discoverable, and testable alongside the rest of the application.

Consistency is still preserved by writing the book mutation and history entry in the same database transaction.

If the database were shared by multiple independent writers and audit completeness had to be guaranteed regardless of the write path, database-level auditing or another event architecture would be worth considering here.

### Transactional history

Book changes and history entries are committed atomically.

This ensures that a successful application-level write cannot exist without its corresponding history record.

Updates and deletions lock the affected book while the previous state is being read and the mutation is applied.

### JSONB instead of JSON

The `changes` column in `history_entries` uses PostgreSQL `JSONB` instead of `JSON`.

The history payload is treated as structured data rather than as an exact textual JSON document. Preserving whitespace, key ordering, or the original byte representation is therefore not important.

`JSONB` also allows efficient querying and indexing of properties inside the history payload if future requirements require filtering or searching specific changes.

`JSON` would mainly be preferable if preserving the exact original JSON text or formatting were important, which is not required for this audit history.

### Explicit SQL

The project uses explicit SQL instead of introducing database views or stored procedures for simple queries.

The current queries are small enough that additional database abstractions would add indirection without providing meaningful reuse.

Dynamic ordering values are restricted through explicit allowlists rather than interpolating unrestricted client values into SQL.

### Soft deletion

Books are soft-deleted rather than physically removed.

This keeps their history available and preserves the foreign-key relationship between a book and its audit entries.

### Full before/after snapshots

Update history stores full before and after snapshots rather than only the properties that changed.

For this service, this keeps the audit representation simple and makes it possible to reconstruct the complete state around an update.

For a larger domain, a more compact change-set representation could be considered.

### Partial update model

PATCH requests are represented separately from the persisted book model.

Optional pointer fields distinguish between a field that was omitted and a field that was explicitly supplied.

The current book is loaded and locked inside the update transaction before the requested fields are applied. This avoids lost updates caused by writing an older full-book representation back to the database.

### Application-layer validation

Input validation is split between transport-level validation and domain/persistence validation.

The HTTP layer validates request shape and individual field constraints.

The storage layer enforces invariants that depend on persisted state, such as referenced authors existing and book titles remaining unique.

This keeps client-facing validation useful without relying solely on request-level checks for database-backed constraints.

### Seed data

Development seed data is kept separate from database migrations.

Migrations define the database schema and invariants, while seed data exists only to make local development and end-to-end testing easier.

The seed currently provides authors because books require existing author references before they can be created through the API.

### Smoke testing

The smoke test uses the public HTTP interface and a real PostgreSQL database.

This provides a lightweight verification that routing, request handling, persistence, transactions, history tracking, and Docker networking work together correctly.

It intentionally does not replace focused Go tests.

### Docker image

The API uses a multi-stage Docker build.

The first stage compiles the Go application, while the runtime image contains only the compiled binary and the minimal runtime environment.

The application runs as a non-root user inside the container.

Database migrations are intentionally run as an explicit deployment step rather than automatically by every API process at startup.

## Production Considerations

The project aims to demonstrate production-oriented application design while remaining appropriately scoped for a small service.

Further concerns for a full production deployment depend on the surrounding platform and deployment environment and may include:

* authentication and authorization;
* metrics and distributed tracing;
* readiness checks;
* TLS termination;
* secret management;
* rate limiting;
* deployment orchestration;
* database backups and recovery;
* CI/CD;
* dependency and container scanning;
* centralized log collection;
* monitoring and alerting.

Development seed data and the smoke-test tooling are intended for development and verification only and are not part of the production application runtime.

These concerns should be implemented according to the environment in which the service is deployed rather than introducing platform-specific infrastructure into the core domain unnecessarily.

## License

This project is licensed under the Apache License 2.0.