# Engineering Contract: Personal Task Tracking API

Artifact produced after Context Loader and Module Registry. Implementation follows
this contract.

## Contract Type

Project Contract

## Mode

Bootstrap Mode

## Problem

Build a personal Task Tracking REST API with Go, Fiber, and SQLite so tasks can
be created, listed, filtered, sorted, updated, and soft-deleted — with clear
architectural layers for learning purposes.

## Context Summary

- **Domain:** Personal task management
- **Task type:** new project
- **Language:** Go
- **Framework:** Fiber
- **Database:** SQLite
- **Existing architecture:** Greenfield
- **Constraints:** No authentication; CRUD only; learning-focused
- **Assumptions:** Single user; local development; env-based config
- **Unknowns (resolved):**
  - All model fields required in schema (`title`, `description`, `status`,
    `priority`, `due_date`, timestamps, soft delete)
  - Status values: `todo`, `in_progress`, `done`
  - Filtering by `status` and `priority`; sorting by `created_at` and `due_date`
  - Soft delete via `deleted_at`
  - Project location: `task-tracker-api/`

## Selected Modules

| Module | Reason |
| --- | --- |
| Core | Always active — KAIOS workflow |
| Context Loader | Task context extracted |
| Module Registry | Module selection documented |
| Engineering Contract | Scope and acceptance criteria |
| Architecture | Layer boundaries and data flow |
| Pattern Engine | Repository and input DTO decisions |
| Testing Engine | Unit and integration coverage |
| Review Engine | Self-review after implementation |
| Learning Engine | Developer learning summary |

## Scope

### In Scope

- Layered architecture: Handler → Service → Repository
- SQLite schema with soft delete
- Task CRUD endpoints:
  - `POST /tasks`
  - `GET /tasks` (filter + sort)
  - `GET /tasks/:id`
  - `PUT /tasks/:id`
  - `DELETE /tasks/:id` (soft delete)
- `GET /health`
- Request validation and consistent JSON errors
- Unit tests (service) and integration tests (handler + SQLite)
- README with run instructions and architecture overview

### Out Of Scope

- Authentication / authorization
- Multi-user / tenant support
- Web or mobile clients
- Pagination, full-text search, tagging
- Docker / deployment pipeline
- Restore endpoint for soft-deleted tasks (v2)
- Audit log

## Requirements

### Functional Requirements

- Create a task with required `title` and optional fields.
- List tasks with optional `status` and `priority` filters.
- Sort tasks by `created_at` or `due_date` (`asc` / `desc`).
- Get a single task by ID; return `404` if missing or soft-deleted.
- Update task fields; return `404` if missing or soft-deleted.
- Soft-delete a task; return `404` if already deleted or missing.
- Reject invalid input with `400`.

### Non-Functional Requirements

- **Security:** Parameterized SQL; input validation; no hardcoded secrets
- **Performance:** Indexes on `status` and `deleted_at`; SQL-level sorting
- **Maintainability:** Clear layer separation; single responsibility per package
- **Testability:** Service tests with mock repository; handler tests with temp DB
- **Scalability:** SQLite sufficient for v1 personal use

## Acceptance Criteria

- [ ] `POST /tasks` with valid body returns `201` and creates a task.
- [ ] Missing or empty `title` returns `400`.
- [ ] `GET /tasks` returns all active (non-deleted) tasks as JSON array.
- [ ] `GET /tasks?status=todo` filters correctly.
- [ ] `GET /tasks?priority=high` filters correctly.
- [ ] `GET /tasks?sort=due_date&order=asc` sorts correctly.
- [ ] `GET /tasks/:id` returns task or `404`.
- [ ] `PUT /tasks/:id` updates task or returns `404` / `400`.
- [ ] `DELETE /tasks/:id` sets `deleted_at` (soft delete), not physical delete.
- [ ] Soft-deleted tasks are excluded from list and get endpoints.
- [ ] Updating a soft-deleted task returns `404`.
- [ ] SQLite path read from configuration (`DB_PATH` env var).
- [ ] At least one service unit test and one handler integration test exist.
- [ ] README documents how to run the API and describes layer structure.

## Architecture Constraints

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
cmd/api/           → bootstrap
internal/config/   → env config
internal/database/ → connection + migrations
internal/task/     → model, service, repository, handler
internal/httpx/    → JSON response helpers
migrations/        → versioned SQL
```

Rules:

- No SQL in handlers.
- No HTTP in repository.
- Service depends on repository interface (consumer-defined).
- Sort parameters mapped via whitelist (no string concatenation).

## Implementation Rules

- Code follows agreed architecture; no premature abstraction.
- Use `internal/` for non-exportable packages.
- Consistent JSON error format across handlers.
- Migrations applied at startup from `migrations/` directory.

## Testing Expectations

- **Unit:** Service create/update validation, not-found paths, soft-delete guards
- **Integration:** Handler CRUD cycle with temp SQLite; filter and sort queries
- **Edge cases:** Empty title, invalid status, non-existent ID
- **Failure cases:** Soft-deleted task returns `404` on get/update

## Review Checklist

- [ ] Architecture boundaries are clear
- [ ] Naming and module cohesion are sensible
- [ ] SOLID and separation of concerns respected
- [ ] SQL injection risks mitigated (parameterized queries, sort whitelist)
- [ ] Tests cover critical paths
- [ ] No over-engineering

## Learning Goals

After this project, the developer should understand:

- Why Handler / Service / Repository are separated
- Fiber's role vs domain logic
- When a repository interface improves testability
- Where validation belongs (handler vs service)
- Soft delete query patterns
- Why contracts precede code in Bootstrap Mode

## Contract Status

**Agreed** — ready for implementation.
