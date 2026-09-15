# go-task-api

A simple REST API for managing tasks, built with Go and PostgreSQL. This
project was built to practice backend fundamentals: designing a REST API,
working with a relational database, structuring a Go project, writing
tests, and running everything with Docker.

## Features

- Full CRUD for tasks: create, read, update, delete
- Task status workflow: `todo`, `in_progress`, `done`
- Input validation with clear, JSON error messages
- Proper HTTP status codes for every scenario (success, bad input, not
  found, server error)
- PostgreSQL storage with parameterized SQL queries (no string
  concatenation, no SQL injection risk)
- Unit and handler-level tests that run without a real database
- One-command startup with Docker Compose

## Tech Stack

- **Go 1.22** (standard `net/http`, using the built-in method + path
  routing introduced in Go 1.22 — no external web framework needed)
- **PostgreSQL 16**
- **lib/pq** — PostgreSQL driver for `database/sql`
- **Docker / Docker Compose**
- Go's built-in `testing` package and `net/http/httptest`

## Project Structure

```
go-task-api/
├── cmd/
│   └── server/
│       └── main.go            # application entry point
├── internal/
│   ├── config/                # environment-based configuration
│   ├── model/                 # Task struct and status constants
│   ├── repository/            # database access (interface + Postgres impl)
│   ├── service/                # validation and business logic
│   ├── handler/                # HTTP handlers (request/response, routing)
│   └── testutil/               # in-memory fake repository, used only in tests
├── migrations/
│   └── 0001_create_tasks_table.sql
├── Dockerfile
├── docker-compose.yml
├── .env.example
├── .gitignore
├── go.mod
└── README.md
```

The project follows a small layered architecture:

- **handler** — parses HTTP requests, writes HTTP responses
- **service** — validates input, applies business rules
- **repository** — talks to PostgreSQL with plain SQL

Each layer only depends on the one below it, through a small interface
(`repository.TaskRepository`), which is what makes it possible to test the
service and handlers without spinning up a real database.

## How to Run

### Option 1 — Docker (recommended)

Requires Docker and Docker Compose.

```bash
cp .env.example .env
docker compose up --build
```

This starts PostgreSQL, creates the `tasks` table automatically (via
`migrations/0001_create_tasks_table.sql`, which Postgres runs on first
startup), and starts the API on `http://localhost:8080`.

### Option 2 — Run locally

Requires Go 1.22+ and a running PostgreSQL instance.

```bash
cp .env.example .env
# edit .env if your local Postgres connection details differ

go mod download
go run ./cmd/server
```

Make sure the `tasks` table exists — either apply
`migrations/0001_create_tasks_table.sql` manually with `psql`, or run the
project once with Docker Compose first (using its own database).

## Environment Variables

| Variable            | Description                                | Default                                                        |
|----------------------|---------------------------------------------|------------------------------------------------------------------|
| `SERVER_PORT`         | Port the HTTP server listens on             | `8080`                                                            |
| `POSTGRES_USER`       | PostgreSQL username (Docker Compose only)   | `postgres`                                                        |
| `POSTGRES_PASSWORD`   | PostgreSQL password (Docker Compose only)   | `postgres`                                                        |
| `POSTGRES_DB`         | PostgreSQL database name (Docker Compose only) | `tasks`                                                        |
| `DATABASE_URL`        | Full Postgres connection string used by the Go app | `postgres://postgres:postgres@localhost:5432/tasks?sslmode=disable` |

See `.env.example` for a ready-to-copy template. No real secrets are
committed to this repository.

## API Endpoints

| Method | Path          | Description            |
|--------|---------------|-------------------------|
| POST   | `/tasks`      | Create a new task       |
| GET    | `/tasks`      | List all tasks          |
| GET    | `/tasks/{id}` | Get a single task by ID |
| PUT    | `/tasks/{id}` | Update a task           |
| DELETE | `/tasks/{id}` | Delete a task           |
| GET    | `/health`     | Health check            |

### Task fields

| Field         | Type   | Notes                                          |
|---------------|--------|-------------------------------------------------|
| `id`          | number | Assigned by the database                        |
| `title`       | string | Required, up to 255 characters                  |
| `description` | string | Optional                                        |
| `status`      | string | One of `todo`, `in_progress`, `done`; defaults to `todo` |
| `created_at`  | string | ISO 8601 timestamp                               |
| `updated_at`  | string | ISO 8601 timestamp                               |

## Example Requests

**Create a task**

```bash
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -d '{"title": "Write README", "description": "Document the API", "status": "in_progress"}'
```

**List tasks**

```bash
curl http://localhost:8080/tasks
```

**Get a single task**

```bash
curl http://localhost:8080/tasks/1
```

**Update a task**

```bash
curl -X PUT http://localhost:8080/tasks/1 \
  -H "Content-Type: application/json" \
  -d '{"title": "Write README", "description": "Document the API", "status": "done"}'
```

**Delete a task**

```bash
curl -X DELETE http://localhost:8080/tasks/1
```

## Example Responses

**201 Created**

```json
{
  "id": 1,
  "title": "Write README",
  "description": "Document the API",
  "status": "in_progress",
  "created_at": "2026-01-10T12:00:00Z",
  "updated_at": "2026-01-10T12:00:00Z"
}
```

**400 Bad Request**

```json
{
  "error": "validation error: title is required"
}
```

**404 Not Found**

```json
{
  "error": "task not found"
}
```

**500 Internal Server Error**

```json
{
  "error": "internal server error"
}
```

## Testing

The `service` and `handler` layers are tested using an in-memory fake
repository (`internal/testutil`), so the full test suite runs without a
real PostgreSQL instance:

```bash
go test ./...
```

Tests cover:

- creating a task (success and validation errors)
- fetching a task by ID (found and not found)
- listing tasks
- updating a task (success and not found)
- deleting a task (success and not found)
- invalid request bodies and invalid IDs at the HTTP layer

The PostgreSQL repository itself (`internal/repository/task_repository.go`)
is exercised end-to-end when the app runs against a real database via
Docker Compose.

## Docker

```bash
docker compose up --build
```

This builds the Go API from `Dockerfile` (a small multi-stage build that
produces a statically linked binary in a minimal Alpine image) and starts
it alongside a PostgreSQL container. The database schema is created
automatically from `migrations/0001_create_tasks_table.sql` on first
startup.

To stop everything:

```bash
docker compose down
```

To also remove the database volume (fresh start):

```bash
docker compose down -v
```

## What I Learned

Building this project helped me practice:

- Structuring a Go backend project into clear layers (handler, service,
  repository) instead of putting everything in `main.go`
- Writing parameterized SQL queries with `database/sql` and understanding
  why string concatenation in SQL is dangerous
- Mapping application errors to correct HTTP status codes
- Using Go interfaces to make business logic testable without a real
  database
- Using Go's built-in `net/http` routing (Go 1.22+) instead of reaching
  for a framework by default
- Containerizing a Go application and PostgreSQL together with Docker
  Compose, including startup ordering and health checks
- Keeping configuration and secrets out of source control with
  `.env.example` and `.gitignore`
