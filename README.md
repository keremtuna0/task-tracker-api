# Personal Task Tracking API

A learning-focused REST API for personal task management, built with Go, Fiber,
and SQLite.

## Features

- Task CRUD with soft delete
- Filter by `status` and `priority`
- Sort by `created_at` or `due_date` (`asc` / `desc`)
- Layered architecture: Handler → Service → Repository
- Parameterized SQL and consistent JSON error responses

## Requirements

- Go 1.25+
- SQLite (via `modernc.org/sqlite`)

## Quick Start

```bash
export DB_PATH=./tasks.db   # Windows: set DB_PATH=.\tasks.db
export PORT=8080            # optional, defaults to 8080

go run ./cmd/api
```

Health check:

```bash
curl http://localhost:8080/health
```

## API Endpoints

| Method | Path | Description |
| --- | --- | --- |
| GET | `/health` | Health check |
| POST | `/tasks` | Create a task |
| GET | `/tasks` | List tasks (filter + sort) |
| GET | `/tasks/:id` | Get one task |
| PUT | `/tasks/:id` | Update a task |
| DELETE | `/tasks/:id` | Soft delete a task |

### Create Task

```bash
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"Learn Go","priority":"high","status":"todo"}'
```

Required: `title`

Optional: `description`, `status` (`todo`, `in_progress`, `done`), `priority`
(`low`, `medium`, `high`), `due_date` (RFC3339 string)

### List Tasks

```bash
curl "http://localhost:8080/tasks?status=todo&priority=high&sort=due_date&order=asc"
```

Query params:

- `status` — filter by status
- `priority` — filter by priority
- `sort` — `created_at` or `due_date` (default: `created_at`)
- `order` — `asc` or `desc` (default: `desc`)

## Architecture

```text
HTTP Handlers (Fiber)
    ↓
Task Service (validation, orchestration)
    ↓
Task Repository (SQLite access)
    ↓
SQLite DB
```

Package layout:

```text
cmd/api/           → bootstrap and route wiring
internal/config/   → env-based configuration
internal/database/ → connection + migrations
internal/task/     → model, service, repository, handler
internal/httpx/    → JSON response helpers
migrations/        → versioned SQL schema
```

Design rules:

- Handlers parse HTTP; no SQL in handlers
- Service owns business validation
- Repository owns data access; no HTTP in repository
- Service depends on a repository interface (consumer-defined)
- Sort fields use a whitelist to prevent SQL injection

## Configuration

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `DB_PATH` | yes | — | SQLite database file path |
| `PORT` | no | `8080` | HTTP listen port |

## Tests

```bash
go test ./...
```

- **Unit tests** (`internal/task/service_test.go`): validation, not-found paths,
  soft-delete guards (mock repository)
- **Integration tests** (`internal/task/handler_test.go`): full CRUD cycle with
  temp SQLite, filters, sort, and soft-delete behavior

## Engineering Contract

Scope, acceptance criteria, and architecture constraints are defined in
[`docs/engineering-contract.md`](docs/engineering-contract.md).
