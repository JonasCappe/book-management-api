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
* Automated tests

## Tech Stack

* Go
* chi
* PostgreSQL 17
* `database/sql`
* `lib/pq`
* `go-playground/validator`
* Swagger / swaggo
* golang-migrate
* Docker Compose
* direnv

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

Cross-cutting concerns such as JSON handling, validation, error responses, pagination, configuration, and database access are separated into dedicated components.

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
├── scripts/                    # Database initialization scripts
├── docker-compose.yml
├── Makefile
└── .env.example
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

UPDATE book

replace author relationships

load resulting state

INSERT history entry

COMMIT
```

If any step fails, the transaction is rolled back.

This means an application-level book mutation cannot successfully commit without its associated history entry also being persisted.

Updates use `SELECT ... FOR UPDATE` to lock the affected book while its previous state is being read and the corresponding history entry is constructed.

## Requirements

For local development:

* Go
* Docker
* Docker Compose
* direnv
* golang-migrate
* swag CLI, when regenerating API documentation
* Make

## Configuration

Copy the example environment file:

```bash
cp .env.example .env
```

The application currently supports the following configuration:

```env
ADDR=":8080"

DB_DSN="postgres://postgres:postgres@localhost:5432/book-management?sslmode=disable"

DB_MAX_OPEN_CONNS=50
DB_MAX_IDLE_CONNS=50
DB_MAX_IDLE_TIME="15m"

POSTGRES_DB="book-management"
POSTGRES_USER="postgres"
POSTGRES_PASSWORD="postgres"
```

Do not use the example database credentials in a production environment.

## Running Locally

### 1. Start PostgreSQL

```bash
make db-up
```

The development environment uses PostgreSQL 17 running through Docker Compose.

### 2. Apply migrations

```bash
make migrate-up
```

Check the current migration version with:

```bash
make migrate-version
```

### 3. Run the API

```bash
make run
```

By default, using the example environment configuration, the API listens on:

```text
http://localhost:8080
```

## Available Make Targets

Display the available development commands:

```bash
make help
```

Common commands include:

```bash
make run
make test

make db-up
make db-down
make db-logs
make db-config

make migrate-up
make migrate-down
make migrate-version
```

Create a new migration:

```bash
make migrate-create name=add_example_column
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

* a non-blank title
* a publication date
* at least one existing author

Duplicate author IDs are rejected by request validation.

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

## Listing Books

The book collection supports pagination, filtering, and ordering.

Example:

```http
GET /v1/books?page=1&page_size=20&title=hobbit&order_by=title&order=asc
```

Supported query parameters:

| Parameter        | Description                             |
| ---------------- | --------------------------------------- |
| `page`           | Page number, starting at 1              |
| `page_size`      | Number of records per page, maximum 100 |
| `title`          | Partial, case-insensitive title filter  |
| `author_id`      | Filter by author                        |
| `published_from` | Minimum publication date                |
| `published_to`   | Maximum publication date                |
| `order_by`       | Sort field                              |
| `order`          | `asc` or `desc`                         |

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

## JSON Request Handling

JSON request bodies are deliberately strict.

The API:

* limits request bodies to 1 MiB
* rejects malformed JSON
* rejects unknown fields
* rejects incorrect value types
* rejects empty request bodies
* rejects requests containing multiple JSON objects

This helps detect client mistakes rather than silently ignoring invalid input.

## Errors

Errors use a consistent JSON representation:

```json
{
  "error": {
    "code": "validation_failed",
    "message": "request validation failed",
    "fields": {
      "title": "title is required"
    }
  }
}
```

Examples of error categories include:

```text
invalid_request
validation_failed
not_found
conflict
internal_error
```

Internal database or implementation errors are not exposed directly to API consumers.

## API Documentation

Interactive Swagger documentation is exposed at:

```text
/swagger/
```

Generated OpenAPI files are stored in:

```text
docs/
```

Regenerate them with:

```bash
make gen-docs
```

## Database Migrations

Migrations are stored under:

```text
cmd/migrate/migrations/
```

Schema changes are applied incrementally rather than modifying an existing migration after it has been introduced.

This repository currently includes migrations covering:

* authors
* books
* history entries
* soft deletion
* publication date and many-to-many book authors

## Testing

Run the full Go test suite with:

```bash
make test
```

Or directly:

```bash
go test ./...
```

Tests cover API helpers, query parsing, history behavior, and persistence-related functionality.

## Design Decisions

### History in the application layer

History generation is implemented in Go rather than through PostgreSQL triggers.

The API owns the book mutation path, so keeping this behavior in the application makes it explicit, discoverable, and testable alongside the rest of the application.

Consistency is still preserved by writing the book mutation and history entry in the same database transaction.

If the database were shared by multiple independent writers and audit completeness has to guaranteed regardless of the write path, database-level auditing or another event architecture would be worth considering here.

### JSONB instead of JSON

The `changes` column in `history_entries` uses PostgreSQL `JSONB` instead of `JSON`.

The history payload is treated as structured data rather than as an exact textual JSON document. Preserving whitespace, key ordering, or the original byte representation is therefore not important.

`JSONB` also allows efficient querying and indexing of properties inside the history payload if future requirements require filtering or searching specific changes.

`JSON` would mainly be preferable if preserving the exact original JSON text or formatting were important, which is not required for this audit history.

### Explicit SQL

The project uses explicit SQL instead of introducing database views or stored procedures for simple queries.

The current queries are small enough that additional database abstractions would add indirection without providing meaningful reuse.

### Soft deletion

Books are soft-deleted rather than physically removed.

This keeps their history available and preserves the foreign-key relationship between a book and its audit entries.

### Full before/after snapshots

Update history stores full before and after snapshots rather than only the properties that changed.

For this service, this keeps the audit representation simple and makes it possible to reconstruct the complete state around an update.

For a larger domain, a more compact change-set representation could be considered.

## Production Considerations

The project aims to demonstrate production-oriented application design while remaining appropriately scoped for a small service.

Further concerns for a full production deployment depend on the surrounding platform and deployment environment and may include:

* authentication and authorization
* structured logging and centralized log collection
* metrics and distributed tracing
* readiness checks
* TLS termination
* secret management
* rate limiting
* deployment orchestration
* database backups and recovery
* CI/CD
* dependency and container scanning
* monitoring and alerting

These concerns should be implemented according to the environment in which the service is deployed rather than introducing platform-specific infrastructure into the core domain unnecessarily.

## License

This project is licensed under the Apache License 2.0.
